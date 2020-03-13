package vmwarevsphere

import (
	"archive/tar"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/rancher/machine/libmachine/log"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/guest"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

func (d *Driver) remoteExec(procman *guest.ProcessManager, arg string) (int64, error) {
	var env []string
	auth := NewAuthFlag(d.SSHUser, d.SSHPassword)
	guestspec := types.GuestProgramSpec{
		ProgramPath:      "/usr/bin/sudo",
		Arguments:        arg,
		WorkingDirectory: "",
		EnvVariables:     env,
	}

	code, err := procman.StartProgram(d.getCtx(), auth.Auth(), &guestspec)
	if err != nil {
		return -1, err
	}

	return code, nil
}

func (d *Driver) getDatastore(spec *types.VirtualMachineConfigSpec) (*object.Datastore, error) {
	var sp *object.StoragePod
	if d.DatastoreCluster != "" {
		log.Infof("Finding datastore cluster %s", d.DatastoreCluster)
		sp, err := d.finder.DatastoreCluster(d.getCtx(), d.DatastoreCluster)
		if err != nil {
			return nil, err
		}

		return d.recommendDatastore(sp, spec)
	}

	if d.Datastore != "" {
		log.Infof("Finding datastore %s", d.Datastore)
		return d.finder.Datastore(d.getCtx(), d.Datastore)
	}

	//nothing set, try default ds cluster then default ds
	log.Infof("Finding default datastore cluster")
	sp, err := d.finder.DefaultDatastoreCluster(d.getCtx())
	if err != nil {
		return nil, err
	}

	if sp != nil {
		return d.recommendDatastore(sp, spec)
	}

	log.Infof("No default datastore cluster, finding default datastore")
	return d.finder.DefaultDatastore(d.getCtx())
}

func (d *Driver) recommendDatastore(sp *object.StoragePod, spec *types.VirtualMachineConfigSpec) (*object.Datastore, error) {
	// Build pod selection spec from config spec
	spref := sp.Folder.Reference()
	podSelectionSpec := types.StorageDrsPodSelectionSpec{
		StoragePod: &spref,
	}

	// Keep list of disks that need to be placed
	var disks []*types.VirtualDisk
	// Collect disks eligible for placement
	for _, deviceConfigSpec := range spec.DeviceChange {
		s := deviceConfigSpec.GetVirtualDeviceConfigSpec()
		if s.Operation != types.VirtualDeviceConfigSpecOperationAdd {
			continue
		}

		if s.FileOperation != types.VirtualDeviceConfigSpecFileOperationCreate {
			continue
		}

		d, ok := s.Device.(*types.VirtualDisk)
		if !ok {
			continue
		}

		podConfigForPlacement := types.VmPodConfigForPlacement{
			StoragePod: spref,
			Disk: []types.PodDiskLocator{
				{
					DiskId:          d.Key,
					DiskBackingInfo: d.Backing,
				},
			},
		}

		podSelectionSpec.InitialVmConfig = append(podSelectionSpec.InitialVmConfig, podConfigForPlacement)
		disks = append(disks, d)
	}

	rp := d.resourcepool.Reference()
	sps := types.StoragePlacementSpec{
		Type:             string(types.StoragePlacementSpecPlacementTypeCreate),
		ResourcePool:     &rp,
		PodSelectionSpec: podSelectionSpec,
		ConfigSpec:       spec,
	}

	c, err := d.getSoapClient()
	if err != nil {
		return nil, err
	}

	srm := object.NewStorageResourceManager(c.Client)
	result, err := srm.RecommendDatastores(d.getCtx(), sps)
	if err != nil {
		return nil, err
	}

	// Use result to pin disks to recommended datastores
	recs := result.Recommendations
	if len(recs) == 0 {
		return nil, fmt.Errorf("no datastore-cluster recommendations")
	}

	ds := recs[0].Action[0].(*types.StoragePlacementAction).Destination

	var mds mo.Datastore
	err = property.DefaultCollector(c.Client).RetrieveOne(d.getCtx(), ds, []string{"name"}, &mds)
	if err != nil {
		return nil, err
	}

	// Apply recommendation to eligible disks
	for _, disk := range disks {
		backing := disk.Backing.(*types.VirtualDiskFlatVer2BackingInfo)
		backing.Datastore = &ds
	}

	return d.finder.Datastore(d.getCtx(), mds.Name)
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

// Make a boot2docker userdata.tar key bundle
func (d *Driver) generateKeyBundle() error {
	log.Debugf("Creating Tar key bundle...")
	magicString := "boot2docker, this is vmware speaking"

	tf, err := os.Create(d.ResolveStorePath("userdata.tar"))
	if err != nil {
		return err
	}
	defer tf.Close()
	var fileWriter = tf

	tw := tar.NewWriter(fileWriter)
	defer tw.Close()

	// magicString first so we can figure out who originally wrote the tar.
	file := &tar.Header{Name: magicString, Size: int64(len(magicString))}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}

	if _, err := tw.Write([]byte(magicString)); err != nil {
		return err
	}

	// .ssh/key.pub => authorized_keys
	file = &tar.Header{Name: ".ssh", Typeflag: tar.TypeDir, Mode: 0700}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}

	pubKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}

	file = &tar.Header{Name: ".ssh/authorized_keys", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}

	if _, err := tw.Write([]byte(pubKey)); err != nil {
		return err
	}

	file = &tar.Header{Name: ".ssh/authorized_keys2", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return err
	}

	if _, err := tw.Write([]byte(pubKey)); err != nil {
		return err
	}

	return nil
}

func (d *Driver) soapLogin() (*govmomi.Client, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s:%d/sdk", d.IP, d.Port))
	if err != nil {
		return nil, err
	}

	u.User = url.UserPassword(d.Username, d.Password)
	c, err := govmomi.NewClient(d.getCtx(), u, true)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (d *Driver) getCtx() context.Context {
	if d.ctx == nil {
		d.ctx = context.Background()
	}

	return d.ctx
}

func (d *Driver) getSoapClient() (*govmomi.Client, error) {
	if d.soap == nil {
		c, err := d.soapLogin()
		if err != nil {
			return nil, err
		}
		d.soap = c
	}

	return d.soap, nil
}

func (d *Driver) getRestLogin(c *vim25.Client) *rest.Client {
	return rest.NewClient(c)
}

func (d *Driver) getUserInfo() *url.Userinfo {
	return url.UserPassword(d.Username, d.Password)
}

func (d *Driver) restLogin(ctx context.Context, c *vim25.Client) (*library.Manager, error) {
	mgr := library.NewManager(rest.NewClient(c))
	ui := url.UserPassword(d.Username, d.Password)
	err := mgr.Login(ctx, ui)
	if err != nil {
		return nil, err
	}

	return mgr, nil
}

func (d *Driver) findFolder() (*object.Folder, error) {
	folders, err := d.datacenter.Folders(d.getCtx())
	if err != nil {
		return nil, err
	}

	if d.Folder != "" {
		p := d.Folder
		if !strings.HasPrefix(p, path.Join(d.Datacenter, "vm")) {
			p = path.Join(folders.VmFolder.InventoryPath, strings.Trim(p, "/"))
		}

		return d.finder.Folder(d.getCtx(), p)
	}

	// root vm folder default
	return folders.VmFolder, nil
}

func (d *Driver) getVAppConfig() *types.VmConfigSpec {
	if d.VAppTransport != "com.vmware.guestInfo" && d.VAppTransport != "iso" {
		return nil
	}

	vApp := types.VmConfigSpec{
		OvfEnvironmentTransport: []string{d.VAppTransport},
	}

	if d.VAppIpAllocationPolicy == "dhcp" ||
		d.VAppIpAllocationPolicy == "fixed" ||
		d.VAppIpAllocationPolicy == "transient" ||
		d.VAppIpAllocationPolicy == "fixedAllocated" {

		if d.VAppIpProtocol != "IPv4" &&
			d.VAppIpProtocol != "IPv6" {
			d.VAppIpProtocol = "IPv4"
		}

		supportedAllocationScheme := "ovfenv"
		if d.VAppIpAllocationPolicy == "dhcp" {
			supportedAllocationScheme = "dhcp"
		}

		vApp.IpAssignment = &types.VAppIPAssignmentInfo{
			SupportedIpProtocol:       []string{d.VAppIpProtocol},
			SupportedAllocationScheme: []string{supportedAllocationScheme},
			IpProtocol:                d.VAppIpProtocol,
			IpAllocationPolicy:        d.VAppIpAllocationPolicy + "Policy",
		}
	}

	for i, prop := range d.VAppProperties {
		v := strings.SplitN(prop, "=", 2)
		key := v[0]
		typ := "string"
		value := ""
		if len(v) > 1 {
			value = v[1]
		}
		if strings.HasPrefix(value, "ip:") {
			typ = value
			value = ""
		} else if strings.HasPrefix(value, "${") &&
			strings.HasSuffix(value, "}") {
			typ = "expression"
		}
		vApp.Property = append(vApp.Property, types.VAppPropertySpec{
			ArrayUpdateSpec: types.ArrayUpdateSpec{
				Operation: types.ArrayUpdateOperationAdd,
			},
			Info: &types.VAppPropertyInfo{
				Key:          int32(i),
				Id:           key,
				Type:         typ,
				DefaultValue: value,
			},
		})
	}

	return &vApp
}

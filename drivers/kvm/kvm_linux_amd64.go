package kvm

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/alexzorin/libvirt-go"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/provider"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
)

const (
	dockerConfigDir = "/var/lib/boot2docker"
	isoFileName     = "boot2docker.iso"
	dnsmasqLeases   = "/var/lib/libvirt/dnsmasq/%s.leases"
	dnsmasqStatus   = "/var/lib/libvirt/dnsmasq/%s.status"

	// TODO
	// - Support multiple NICs
	// - Support bridged networks via static IP configuration
	domainXML = `<domain type='kvm'>
  <name>%s</name> <memory unit='M'>%d</memory>
  <vcpu>%d</vcpu>
  <features><acpi/><apic/><pae/></features>
  <os>
    <type>hvm</type>
    <boot dev='cdrom'/>
    <boot dev='hd'/>
    <bootmenu enable='no'/>
  </os>
  <devices>
    <disk type='file' device='cdrom'>
      <source file='%s'/>
      <target dev='hdc' bus='ide'/>
      <readonly/>
    </disk>
    <disk type='file' device='disk'>
      <source file='%s'/>
      <target dev='hda' bus='ide'/>
    </disk>
    <graphics type='vnc' autoport='yes' listen='127.0.0.1'>
      <listen type='address' address='127.0.0.1'/>
    </graphics>
    <interface type='network'>
      <source network='%s'/>
    </interface>
  </devices>
</domain>`
)

type Driver struct {
	MachineName      string
	SSHUser          string
	Memory           int
	DiskSize         int
	VCPU             int
	Network          string
	Boot2DockerURL   string
	CaCertPath       string
	PrivateKeyPath   string
	SwarmMaster      bool
	SwarmHost        string
	SwarmDiscovery   string
	storePath        string
	connectionString string
	conn             *libvirt.VirConnection
	VM               *libvirt.VirDomain
	vmLoaded         bool
}

type CreateFlags struct {
	Memory         *int
	DiskSize       *int
	VCPU           *int
	Network        *string
	Boot2DockerURL *string
}

func init() {
	log.Debugf("kvm driver init")
	drivers.Register("kvm", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		/*
		 * Can't support this at present due to filesystem assumptions
		 * If we can figure out how to copy the disk image up
		 * to the remote system, then we could support remote libvirt
		 * instances
		 */
		/*
			cli.StringFlag{
				Name:  "kvm-connection",
				Usage: "The libvirt connection string",
				Value: "qemu:///system",
			},
		*/
		cli.IntFlag{
			Name:  "kvm-memory",
			Usage: "Size of memory for host in MB",
			Value: 1024,
		},
		cli.IntFlag{
			Name:  "kvm-disk-size",
			Usage: "Size of disk for host in MB",
			Value: 20000,
		},
		cli.IntFlag{
			Name:  "kvm-vcpu",
			Usage: "Number of virtual CPUs",
			Value: 1,
		},
		// TODO - support for multiple networks
		cli.StringFlag{
			Name:  "kvm-network",
			Usage: "Name of network to connect to",
			Value: "default",
		},
		// TODO - Seems like a candidate for a shared flag...
		cli.StringFlag{
			EnvVar: "KVM_BOOT2DOCKER_URL",
			Name:   "kvm-boot2docker-url",
			Usage:  "The URL of the boot2docker image. Defaults to the latest available version",
			Value:  "",
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	d := Driver{MachineName: machineName, storePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey}
	conn, err := libvirt.NewVirConnection(d.connectionString)
	if err != nil {
		log.Warnf("Failed to connect: %s", err)
		return nil, err
	}
	d.conn = &conn
	return &d, nil
}

func (d *Driver) AuthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) DeauthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) GetSSHPort() (int, error) {
	return 22, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "docker"
	}

	return d.SSHUser
}

func (d *Driver) GetProviderType() provider.ProviderType {
	return provider.Local
}

func (d *Driver) DriverName() string {
	return "kvm"
}

func (d *Driver) GetURL() (string, error) {
	log.Debugf("GetURL called")
	ip, err := d.GetIP()
	if err != nil {
		log.Warnf("Failed to get IP: %s", err)
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	log.Debugf("SetConfigFromFlags aclled")
	d.Memory = flags.Int("kvm-memory")
	d.DiskSize = flags.Int("kvm-disk-size")
	d.VCPU = flags.Int("kvm-vcpu")
	d.Network = flags.String("kvm-network")
	d.Boot2DockerURL = flags.String("kvm-boot2docker-url")

	//d.connectionString = flags.String("kvm-connection")
	d.connectionString = "qemu:///system"

	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = "docker"
	return nil
}

func (d *Driver) validateNetwork(name string) error {
	log.Debugf("Validating network %s", name)
	network, err := d.conn.LookupNetworkByName(name)
	if err != nil {
		log.Warnf("Unable to locate network %s", name)
		return err
	}
	xmldoc, err := network.GetXMLDesc(0)
	if err != nil {
		return err
	}
	// XML structure:
	//  <network>
	//      ...
	//      <forward mode='nat'>
	//      <ip address='a.b.c.d' netmask='255.255.255.0'>
	//          <dhcp>
	//              <range start='a.b.c.d' end='w.x.y.z'/>
	//          </dhcp>
	type Forward struct {
		Mode string `xml:"mode,attr"`
	}
	type Ip struct {
		Address string `xml:"address,attr"`
		Netmask string `xml:"netmask,attr"`
	}
	type Network struct {
		Forward Forward `xml:"forward"`
		Ip      Ip      `xml:"ip"`
	}

	var nw Network
	err = xml.Unmarshal([]byte(xmldoc), &nw)
	if err != nil {
		return err
	}
	switch nw.Forward.Mode {
	case "nat":
		fallthrough
	case "route":
		fallthrough
	case "": // No forwarding
		if nw.Ip.Address == "" {
			log.Warnf("Network doesn't have valid IP range")
			return errors.New("Unsupported libvirt network mode")
		}
	case "bridge":
		log.Infof("Network %s bridged", name)
		/*
		 * TODO
		 *  There isn't really a silver bullet for bridged networks.
		 *  Since the DHCP server is most likely not on this box,
		 *  we can't look up the IP address, and our ARP
		 *  table probably wont have the MAC listed.  To support
		 *  this we can ask the user to give us the IP.
		 */
		return errors.New("Unsupported libvirt network mode")
	default:
		log.Infof("Network %s is not support, but %s instead", name, nw.Forward.Mode)
		return errors.New("Unsupported libvirt network mode")
	}
	return nil
}

func (d *Driver) PreCreateCheck() error {
	// We could look at d.conn.GetCapabilities()
	// parse the XML, and look for hypervisors we care about

	// TODO might want to check minimum version
	_, err := d.conn.GetLibVersion()
	if err != nil {
		log.Warnf("Unable to get libvirt version")
		return err
	}
	err = d.validateNetwork(d.Network)
	if err != nil {
		return err
	}
	// Others...?
	return nil
}

func (d *Driver) Create() error {
	var (
		err     error
		isoURL  string
		isoDest string
	)
	imgPath := utils.GetMachineCacheDir()
	isoFilename := "boot2docker.iso"
	commonIsoPath := filepath.Join(imgPath, "boot2docker.iso")
	// just in case boot2docker.iso has been manually deleted
	if _, err := os.Stat(imgPath); os.IsNotExist(err) {
		if err := os.Mkdir(imgPath, 0700); err != nil {
			return err
		}

	}

	// Libvirt typically runs as a deprivileged service account and
	// needs the execute bit set for directories that contain disks
	for dir := d.storePath; dir != "/"; dir = filepath.Dir(dir) {
		log.Debugf("Verifying executable bit set on %s", dir)
		info, err := os.Stat(dir)
		if err != nil {
			return err
		}
		mode := info.Mode()
		if mode&0001 != 1 {
			log.Debugf("Setting executable bit set on %s", dir)
			mode |= 0001
			os.Chmod(dir, mode)
		}
	}

	// Download the boot2docker ISO
	// TODO - this should be a utility (lifted verbatim from virtualbox.go)
	b2dutils := utils.NewB2dUtils("", "")
	if d.Boot2DockerURL != "" {
		isoURL = d.Boot2DockerURL
		log.Infof("Downloading %s from %s...", isoFilename, isoURL)
		if err := b2dutils.DownloadISO(d.storePath, isoFilename, isoURL); err != nil {
			return err

		}

	} else {
		// todo: check latest release URL, download if it's new
		// until then always use "latest"
		isoURL, err = b2dutils.GetLatestBoot2DockerReleaseURL()
		if err != nil {
			log.Warnf("Unable to check for the latest release: %s", err)
		}
		if _, err := os.Stat(commonIsoPath); os.IsNotExist(err) {
			log.Infof("Downloading %s to %s...", isoFilename, commonIsoPath)
			if err := b2dutils.DownloadISO(imgPath, isoFilename, isoURL); err != nil {

				return err
			}

		}
		isoDest = filepath.Join(d.storePath, isoFilename)
		if err := utils.CopyFile(commonIsoPath, isoDest); err != nil {
			return err
		}
	}

	log.Debugf("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	log.Debugf("Creating VM data disk...")
	if err := d.generateDiskImage(d.DiskSize); err != nil {
		return err
	}

	log.Debugf("Defining VM...")
	// TODO Needs love for other tunables users might want to tweak
	xml := fmt.Sprintf(domainXML, d.MachineName, d.Memory, d.VCPU,
		isoDest, d.diskPath(), d.Network)
	vm, err := d.conn.DomainDefineXML(xml)
	if err != nil {
		log.Warnf("Failed to create the VM: %s", err)
		return err
	}
	d.VM = &vm
	d.vmLoaded = true

	return d.Start()
}

func (d *Driver) Start() error {
	log.Debugf("Starting VM %s", d.MachineName)
	d.validateVMRef()
	err := d.VM.Create()
	if err != nil {
		log.Warnf("Failed to start: %s", err)
		return err
	}

	for i := 0; i < 90; i++ {
		time.Sleep(time.Second)
		ip, _ := d.GetIP()
		if ip != "" {
			// Add a second to let things settle
			time.Sleep(time.Second)
			return nil
		}
		log.Debugf("Waiting for the VM to come up... %d", i)
	}
	log.Warnf("Unable to determine VM's IP address, did it fail to boot?")
	return err
}

func (d *Driver) Stop() error {
	log.Debugf("Stopping VM %s", d.MachineName)
	d.validateVMRef()
	s, err := d.GetState()
	if err != nil {
		return err
	}

	if s != state.Stopped {
		err := d.VM.DestroyFlags(libvirt.VIR_DOMAIN_DESTROY_GRACEFUL)
		if err != nil {
			log.Warnf("Failed to gracefully shutdown VM")
			return err
		}
		for i := 0; i < 90; i++ {
			time.Sleep(time.Second)
			s, _ := d.GetState()
			log.Debugf("VM state: %s", s)
			if s == state.Stopped {
				return nil
			}
		}
		return errors.New("VM Failed to gracefully shutdown, try the kill command")
	}
	return nil
}

func (d *Driver) Remove() error {
	log.Debugf("Removing VM %s", d.MachineName)
	d.validateVMRef()
	// Note: If we switch to qcow disks instead of raw the user
	//       could take a snapshot.  If you do, then Undefine
	//       will fail unless we nuke the snapshots first
	d.VM.Destroy() // Ignore errors
	return d.VM.Undefine()
}

func (d *Driver) Restart() error {
	log.Debugf("Restarting VM %s", d.MachineName)
	if err := d.Stop(); err != nil {
		return err
	}
	return d.Start()
}

func (d *Driver) Kill() error {
	log.Debugf("Killing VM %s", d.MachineName)
	d.validateVMRef()
	return d.VM.Destroy()
}

func (d *Driver) GetState() (state.State, error) {
	log.Debugf("Getting current state...")
	d.validateVMRef()
	states, err := d.VM.GetState()
	if err != nil {
		return state.None, err
	}
	switch states[0] {
	case libvirt.VIR_DOMAIN_NOSTATE:
		return state.None, nil
	case libvirt.VIR_DOMAIN_RUNNING:
		return state.Running, nil
	case libvirt.VIR_DOMAIN_BLOCKED:
		// TODO - Not really correct, but does it matter?
		return state.Error, nil
	case libvirt.VIR_DOMAIN_PAUSED:
		return state.Paused, nil
	case libvirt.VIR_DOMAIN_SHUTDOWN:
		return state.Stopped, nil
	case libvirt.VIR_DOMAIN_CRASHED:
		return state.Error, nil
	case libvirt.VIR_DOMAIN_PMSUSPENDED:
		return state.Saved, nil
	case libvirt.VIR_DOMAIN_SHUTOFF:
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) validateVMRef() {
	if !d.vmLoaded {
		log.Debugf("Fetching VM...")
		vm, err := d.conn.LookupDomainByName(d.MachineName)
		if err != nil {
			log.Warnf("Failed to fetch machine")
		} else {
			d.VM = &vm
			d.vmLoaded = true
		}
	}
}

// This implementation is specific to default networking in libvirt
// with dnsmasq
func (d *Driver) getMAC() (string, string, error) {
	d.validateVMRef()
	xmldoc, err := d.VM.GetXMLDesc(0)
	if err != nil {
		return "", "", err
	}
	// XML structure:
	//  <domain>
	//      ...
	//      <devices>
	//          ...
	//          <interface type='network'>
	//              ...
	//              <mac address='52:54:00:d2:3f:ba'/>
	//              ...
	//          </interface>
	//          ...
	type Mac struct {
		Address string `xml:"address,attr"`
	}
	type Source struct {
		Network string `xml:"network,attr"`
	}
	type Interface struct {
		Type   string `xml:"type,attr"`
		Mac    Mac    `xml:"mac"`
		Source Source `xml:"source"`
	}
	type Devices struct {
		Interfaces []Interface `xml:"interface"`
	}
	type Domain struct {
		Devices Devices `xml:"devices"`
	}

	var dom Domain
	err = xml.Unmarshal([]byte(xmldoc), &dom)
	if err != nil {
		return "", "", err
	}
	for _, iface := range dom.Devices.Interfaces {
		log.Debugf("VM MAC: %s", iface.Mac.Address)
		return iface.Mac.Address, iface.Source.Network, nil
	}
	log.Infof("Unable to locate MAC address")
	return "", "", nil
}

func (d *Driver) getIPByMACFromLeaseFile(mac string, network string) (string, error) {
	leaseFile := fmt.Sprintf(dnsmasqLeases, network)
	data, err := ioutil.ReadFile(leaseFile)
	if err != nil {
		log.Debugf("Failed to retrieve dnsmasq leases from %s", leaseFile)
		return "", err
	}
	for lineNum, line := range strings.Split(string(data), "\n") {
		if len(line) == 0 {
			continue
		}
		entries := strings.Split(line, " ")
		if len(entries) < 3 {
			log.Warnf("Malformed dnsmasq line %d", lineNum+1)
			return "", errors.New("Malformed dnsmasq file")
		}
		if strings.ToLower(entries[1]) == strings.ToLower(mac) {
			log.Debugf("IP address: %s", entries[2])
			return entries[2], nil
		}
	}
	log.Debug("Unable to locate IP address for MAC %s", mac)
	return "", nil
}

func (d *Driver) getIPByMacFromSettings(mac string, network_name string) (string, error) {
	log.Debugf("Looking for MAC in settings")
	network, err := d.conn.LookupNetworkByName(network_name)
	if err != nil {
		log.Warnf("Failed to find network: %s", err)
		return "", err
	}
	bridge_name, err := network.GetBridgeName()
	if err != nil {
		log.Warnf("Failed to get network bridge: %s", err)
		return "", err
	}
	statusFile := fmt.Sprintf(dnsmasqStatus, bridge_name)
	data, err := ioutil.ReadFile(statusFile)
	type Lease struct {
		Ip_address  string `json:"ip-address"`
		Mac_address string `json:"mac-address"`
		/* Other unused fields omitted */
	}
	var s []Lease

	err = json.Unmarshal(data, &s)
	if err != nil {
		log.Warnf("Failed to decode dnsmasq lease status: %s", err)
		return "", err
	}
	for _, value := range s {
		if strings.ToLower(value.Mac_address) == strings.ToLower(mac) {
			log.Debugf("IP address: %s", value.Ip_address)
			return value.Ip_address, nil
		}
	}
	log.Debugf("Couldn't find a match")
	return "", nil
}

func (d *Driver) GetIP() (string, error) {
	log.Debugf("Getting IP...")

	mac, network, err := d.getMAC()
	if err != nil {
		return "", err
	}
	/*
	 * TODO - Figure out what version of libvirt changed behavior and
	 *        be smarter about selecting which algorithm to use
	 */
	ip, err := d.getIPByMACFromLeaseFile(mac, network)
	if ip == "" {
		ip, err = d.getIPByMacFromSettings(mac, network)
	}
	return ip, err
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) diskPath() string {
	return filepath.Join(d.storePath, "disk.img")
}

// TODO - this should be a utility (lifted from virtualbox.go)
// Make a boot2docker VM disk image.
func (d *Driver) generateDiskImage(size int) error {
	log.Debugf("Creating %d MB hard disk image...", size)

	magicString := "boot2docker, please format-me"

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// magicString first so the automount script knows to format the disk
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
	if err := tw.Close(); err != nil {
		return err
	}
	raw := bytes.NewReader(buf.Bytes())
	return createDiskImage(d.diskPath(), size, raw)
}

// createDiskImage makes a disk image at dest with the given size in MB. If r is
// not nil, it will be read as a raw disk image to convert from.
func createDiskImage(dest string, size int, r io.Reader) error {
	// Convert a raw image from stdin to the dest VMDK image.
	sizeBytes := int64(size) << 20 // usually won't fit in 32-bit int (max 2GB)
	f, err := os.Create(dest)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, r)
	if err != nil {
		return err
	}
	// Rely on seeking to create a sparse raw file for qemu
	f.Seek(sizeBytes-1, 0)
	f.Write([]byte{0})
	return f.Close()
}

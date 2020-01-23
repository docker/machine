/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwarevsphere

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/rancher/machine/libmachine/drivers"
	"github.com/rancher/machine/libmachine/log"
	"github.com/rancher/machine/libmachine/mcnutils"
	"github.com/rancher/machine/libmachine/ssh"
	"github.com/rancher/machine/libmachine/state"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"golang.org/x/net/context"
)

const (
	// dockerBridgeIP is the default IP address of the docker0 bridge.
	dockerBridgeIP = "172.17.0.1"
	isoFilename    = "boot2docker.iso"
	// B2DUser is the guest User for
	// tools login
	B2DUser = "docker"
	// B2DPass is the guest Pass for tools login
	B2DPass      = "tcuser"
	B2DUserGroup = "staff"
)

type Driver struct {
	*drivers.BaseDriver
	Memory         int
	DiskSize       int
	CPU            int
	ISO            string
	Boot2DockerURL string
	CPUS           int
	MachineId      string

	IP                     string
	Port                   int
	Username               string
	Password               string
	Network                string
	Networks               []string
	Tags                   []string
	CustomAttributes       []string
	Datastore              string
	DatastoreCluster       string
	Datacenter             string
	Folder                 string
	Pool                   string
	HostSystem             string
	CfgParams              []string
	CloudInit              string
	CloudConfig            string
	VAppIpProtocol         string
	VAppIpAllocationPolicy string
	VAppTransport          string
	VAppProperties         []string
	CreationType           string
	ContentLibrary         string
	CloneFrom              string
	SSHPassword            string
	SSHUserGroup           string

	vms          map[string]*object.VirtualMachine
	soap         *govmomi.Client
	ctx          context.Context
	finder       *find.Finder
	datacenter   *object.Datacenter
	networks     map[string]object.NetworkReference
	hostsystem   *object.HostSystem
	resourcepool *object.ResourcePool
}

const (
	driverName          = "vmwarevsphere"
	defaultSSHUser      = B2DUser
	defaultSSHPass      = B2DPass
	defaultSSHUserGroup = B2DUserGroup
	defaultCpus         = 2
	defaultMemory       = 2048
	defaultDiskSize     = 20480
	defaultSDKPort      = 443

	creationTypeVM      = "vm"
	creationTypeTmpl    = "template"
	creationTypeLibrary = "library"
	creationTypeLegacy  = "legacy"
)

func NewDriver(hostName, storePath string) drivers.Driver {
	d := &Driver{
		CPUS:         defaultCpus,
		Memory:       defaultMemory,
		DiskSize:     defaultDiskSize,
		SSHPassword:  defaultSSHPass,
		SSHUserGroup: defaultSSHUserGroup,
		Port:         defaultSDKPort,
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
			SSHPort:     drivers.DefaultSSHPort,
		},
		ctx:      context.Background(),
		vms:      make(map[string]*object.VirtualMachine),
		networks: make(map[string]object.NetworkReference),
	}
	return d
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = defaultSSHUser
	}

	return d.SSHUser
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

func (d *Driver) GetMachineId() (string, error) {
	if d.MachineId != "" {
		return d.MachineId, nil
	}

	vm, err := d.fetchVM(d.MachineName)
	if err != nil {
		return "", err
	}

	d.MachineId = vm.UUID(d.getCtx())
	return d.MachineId, nil
}

func (d *Driver) GetIP() (string, error) {
	if d.IPAddress != "" {
		return d.IPAddress, nil
	}

	status, err := d.GetState()
	if err != nil {
		return "", err
	}

	if status != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}

	vm, err := d.fetchVM(d.MachineName)
	if err != nil {
		return "", err
	}

	configuredMacIPs, err := vm.WaitForNetIP(d.getCtx(), false)
	if err != nil {
		return "", err
	}

	for _, ips := range configuredMacIPs {
		if len(ips) >= 0 {
			// Prefer IPv4 address, but fall back to first/IPv6
			preferredIP := ips[0]
			for _, ip := range ips {
				// In addition to non IPv4 addresses, try to filter
				// out link local addresses and the default address of
				// the Docker0 bridge
				netIP := net.ParseIP(ip)
				if netIP.To4() != nil && netIP.IsGlobalUnicast() && !netIP.Equal(net.ParseIP(dockerBridgeIP)) {
					preferredIP = ip
					break
				}
			}
			d.IPAddress = preferredIP //cache
			return preferredIP, nil
		}
	}

	return "", errors.New("No IP despite waiting for one - check DHCP status")
}

func (d *Driver) GetState() (state.State, error) {

	// Create context
	c, err := d.getSoapClient()
	if err != nil {
		return state.None, err
	}

	vm, err := d.fetchVM(d.MachineName)
	if err != nil {
		return state.None, err
	}

	var mvm mo.VirtualMachine

	err = c.RetrieveOne(d.getCtx(), vm.Reference(), nil, &mvm)
	if err != nil {
		return state.None, nil
	}

	s := mvm.Summary

	if strings.Contains(string(s.Runtime.PowerState), "poweredOn") {
		return state.Running, nil
	} else if strings.Contains(string(s.Runtime.PowerState), "poweredOff") {
		return state.Stopped, nil
	}
	return state.None, nil
}

// PreCreateCheck checks that the machine creation process can be started safely.
func (d *Driver) PreCreateCheck() error {
	log.Debug("Connecting to vSphere for pre-create checks...")

	err := d.preCreate()
	if err != nil {
		return err
	}

	_, err = d.findFolder()
	if err != nil {
		return err
	}

	if d.CreationType == "clone" && d.ContentLibrary == "" {
		if _, err := d.fetchVM(d.CloneFrom); err != nil {
			return fmt.Errorf("Error finding vm or template to clone: %s", err)
		}
	}

	// TODO: if the user has both the VSPHERE_NETWORK defined and adds --vmwarevsphere-network
	//       both are used at the same time - probably should detect that and remove the one from ENV
	if len(d.Networks) == 0 {
		// machine assumes there will be a network
		d.Networks = append(d.Networks, "VM Network")
	}
	for _, netName := range d.Networks {
		if _, err := d.finder.NetworkOrDefault(d.getCtx(), netName); err != nil {
			return err
		}
	}
	// d.Network needs to remain a string to cope with existing machines :/
	d.Network = d.Networks[0]

	var hs *object.HostSystem
	if d.HostSystem != "" {
		var err error
		hs, err = d.finder.HostSystemOrDefault(d.getCtx(), d.HostSystem)
		if err != nil {
			return err
		}
	}

	// ResourcePool
	if d.Pool != "" {
		// Find specified Resource Pool
		if _, err := d.finder.ResourcePool(d.getCtx(), d.Pool); err != nil {
			return err
		}
	} else if hs != nil {
		// Pick default Resource Pool for Host System
		if _, err := hs.ResourcePool(d.getCtx()); err != nil {
			return err
		}
	} else {
		// Pick the default Resource Pool for the Datacenter.
		if _, err := d.finder.DefaultResourcePool(d.getCtx()); err != nil {
			return err
		}
	}

	return nil
}

func (d *Driver) Create() error {
	log.Infof("Generating SSH Keypair...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	if err := d.preCreate(); err != nil {
		return err
	}

	switch d.CreationType {
	case "legacy":
		log.Infof("Creating VM...")
		b2dutils := mcnutils.NewB2dUtils(d.StorePath)
		if err := b2dutils.CopyIsoToMachineDir(d.Boot2DockerURL, d.MachineName); err != nil {
			return err
		}
		return d.createLegacy()
	case "library":
		log.Infof("Creating VM from /%s/%s...", d.ContentLibrary, d.CloneFrom)
		return d.createFromLibraryName()
	case "vm", "template":
		log.Infof("Cloning VM from VM or Template: %s...", d.CloneFrom)
		return d.createFromVmName()
	default:
		log.Infof("Unable to perform any actions, change flags and try again")
		return nil
	}
}

func (d *Driver) Start() error {
	machineState, err := d.GetState()
	if err != nil {
		return err
	}

	switch machineState {
	case state.Running:
		log.Infof("VM %s has already been started", d.MachineName)
		return nil
	case state.Stopped:
		// TODO add transactional or error handling in the following steps
		vm, err := d.fetchVM(d.MachineName)
		if err != nil {
			return err
		}

		task, err := vm.PowerOn(d.getCtx())
		if err != nil {
			return err
		}

		_, err = task.WaitForResult(d.getCtx(), nil)
		if err != nil {
			return err
		}

		log.Infof("Waiting for VMware Tools to come online...")
		if _, err = d.GetIP(); err != nil {
			return err
		}
		if _, err = d.GetMachineId(); err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) Stop() error {
	vm, err := d.fetchVM(d.MachineName)
	if err != nil {
		return err
	}

	if err := vm.ShutdownGuest(d.getCtx()); err != nil {
		return err
	}

	d.IPAddress = ""

	return nil
}

func (d *Driver) Restart() error {
	if err := d.Stop(); err != nil {
		return err
	}

	// Check for 120 seconds for the machine to stop
	for i := 1; i <= 60; i++ {
		machineState, err := d.GetState()
		if err != nil {
			return err
		}
		if machineState == state.Running {
			log.Debugf("Not there yet %d/%d", i, 60)
			time.Sleep(2 * time.Second)
			continue
		}
		if machineState == state.Stopped {
			break
		}
	}

	machineState, err := d.GetState()
	if err != nil {
		return err
	}

	// If the VM is still running after 120 seconds just kill it.
	if machineState == state.Running {
		if err = d.Kill(); err != nil {
			return fmt.Errorf("can't stop VM: %s", err)
		}
	}

	return d.Start()
}

func (d *Driver) Kill() error {
	vm, err := d.fetchVM(d.MachineName)
	if err != nil {
		return err
	}

	task, err := vm.PowerOff(d.getCtx())
	if err != nil {
		return err
	}

	_, err = task.WaitForResult(d.getCtx(), nil)
	if err != nil {
		return err
	}

	d.IPAddress = ""

	return nil
}

func (d *Driver) Remove() error {
	if d.MachineId == "" {
		//no guid from config, nothing in vsphere to delete
		return nil
	}

	if err := d.preCreate(); err != nil {
		return err
	}

	c, err := d.getSoapClient()
	if err != nil {
		return err
	}

	vm, err := d.fetchVM(d.MachineName)
	if err != nil {
		return err
	}

	if vm.UUID(d.getCtx()) != d.MachineId {
		return fmt.Errorf("Machine Id mismatch trying to delete %s but config has %s", vm.UUID(d.getCtx()), d.MachineId)
	}

	machineState, err := d.GetState()
	if err != nil {
		return err
	}
	if machineState == state.Running {
		if err = d.Kill(); err != nil {
			return fmt.Errorf("can't stop VM: %s", err)
		}
	}

	ds, err := d.getVmDatastore(vm)
	if err != nil {
		return err
	}

	var task *object.Task
	// Remove B2D Iso from VM folder
	if d.CreationType == creationTypeLegacy {
		m := object.NewFileManager(c.Client)
		task, err = m.DeleteDatastoreFile(d.getCtx(), ds.Path(fmt.Sprintf("%s/%s", d.MachineName, isoFilename)), d.datacenter)
		if err != nil {
			return err
		}

		if err = task.Wait(d.getCtx()); err != nil {
			if types.IsFileNotFound(err) {
				// Ignore error
				return nil
			}
		}
	} else {
		if err = d.removeCloudInitIso(vm, d.datacenter, ds); err != nil {
			return nil
		}
	}

	task, err = vm.Destroy(d.getCtx())
	if err != nil {
		return err
	}

	_, err = task.WaitForResult(d.getCtx(), nil)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Upgrade() error {
	return fmt.Errorf("upgrade is not supported for vsphere driver at this moment")
}

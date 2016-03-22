/*
 * docker-machine driver for VMware's photoncontroller
 * photon-controller: http://vmware.github.io/photon-controller/
 *
 * This version is currently just a stub.
 */

package photoncontroller

import (
	"fmt"
	"time"
	"net"
	"strconv"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

const (
	driverName = "photoncontroller"
	defaultEndpoint = "https://192.0.2.2"
	defaultAuthEndpoint = "10.146.64.236"
	defaultSSHUser = "root"
	defaultSSHKeyPath = "/ssh/id_rsa"
	maxRetries = 150
)

type Driver struct {
	*drivers.BaseDriver
	Name           string
	Tenant         string
	Project        string
	VMFlavor       string
	DiskFlavor     string
	DiskName       string
	Image          string
	VMId           string
	PhotonEndpoint string
}

type NotLoadable struct {
	Name string
}

func (e NotLoadable) Error() string {
	return fmt.Sprintf("Driver %q not found. Do you have the plugin binary accessible in your PATH?", e.Name)
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		Tenant:         "TenantsName",
		Project:        "626ce36a-8d61-4b69-8880-81829aac7b8c",
		VMFlavor:       "VMFlavor",
		DiskFlavor:     "DiskFlavor",
		DiskName:       "disk-10",
		Image:          "707f566f-c3d0-44b3-8eb6-e66a087f46dc",
		PhotonEndpoint: defaultEndpoint,
		BaseDriver:     &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
			SSHPort:     22,
			SSHKeyPath:  defaultSSHKeyPath,
		},
	}
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "photon-endpoint",
			Usage:  "Photon Controller endpoint in format: http://<ip address>:<port>",
			Value:  defaultEndpoint,
			EnvVar: "PHOTON_ENDPOINT",
		},
		mcnflag.StringFlag{
			Name:   "photon-tenant",
			Usage:  "Tenant name",
			EnvVar: "PHOTON_TENANT",
		},
		mcnflag.StringFlag{
			Name:   "photon-project",
			Usage:  "Project Id",
			EnvVar: "PHOTON_PROJECT",
		},
		mcnflag.StringFlag{
			Name:   "photon-vmflavor",
			Usage:  "VM flavor name",
			EnvVar: "PHOTON_VM_FLAVOR",
		},
		mcnflag.StringFlag{
			Name:   "photon-diskflavor",
			Usage:  "Disk flavor name",
			EnvVar: "PHOTON_DISK_FLAVOR",
		},
		mcnflag.StringFlag{
			Name:   "photon-diskname",
			Usage:  "Disk name",
			EnvVar: "PHOTON_DISK_Name",
		},
		mcnflag.StringFlag{
			Name:   "photon-image",
			Usage:  "Image Id",
			EnvVar: "PHOTON_IMAGE_ID",
		},
		mcnflag.StringFlag{
			Name:   "photon-ssh-keypath",
			Usage:  "SSH key path",
			Value:  defaultSSHKeyPath,
			EnvVar: "PHOTON_SSH_KEYPATH",
		},
		mcnflag.StringFlag{
			Name:   "photon-ssh-user",
			Usage:  "SSH user",
			Value:  defaultSSHUser,
			EnvVar: "PHOTON_SSH_USER",
		},
	}
}

func (d *Driver) getClient() *photon.Client {
	return photon.NewClient(d.PhotonEndpoint, defaultAuthEndpoint, nil)
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
	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, strconv.Itoa(engine.DefaultPort))), nil
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) GetIP() (string, error) {
	return d.IPAddress, nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHKeyPath() string {
	return d.SSHKeyPath
}

func (d *Driver) GetSSHPort() (int, error) {
	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	return d.SSHUser
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Tenant = flags.String("photon-tenant")
	d.Project = flags.String("photon-project")
	d.VMFlavor = flags.String("photon-vmflavor")
	d.DiskFlavor = flags.String("photon-diskflavor")
	d.DiskName = flags.String("photon-diskname")
	d.Image = flags.String("photon-image")
	d.PhotonEndpoint = flags.String("photon-endpoint")
	d.SSHKeyPath = flags.String("photon-ssh-keypath")
	d.SSHUser = flags.String("photon-ssh-user")
	d.SSHPort = 22
	d.SetSwarmConfigFromFlags(flags)

	return nil
}

func (d *Driver) Create() error {
	vmSpec := &photon.VmCreateSpec {
		Name:          d.MachineName,
		Flavor:        d.VMFlavor,
		SourceImageID: d.Image,
		AttachedDisks: []photon.AttachedDisk{
			photon.AttachedDisk{
				CapacityGB: 1,
				Flavor:     d.DiskFlavor,
				Kind:       "ephemeral-disk",
				Name:       d.DiskName,
				State:      "STARTED",
				BootDisk:   true,
			},
		},
	}
	client := d.getClient()

	// Creating VM task
	vmCreateTask, err := client.Projects.CreateVM(d.Project, vmSpec)
	if (err != nil) {
		return fmt.Errorf("Error creating VM Task: ", err)
	}

	// Waiting for create VM task completion
	vmCreateTask, err = client.Tasks.Wait(vmCreateTask.ID)
	if (err != nil) {
		return fmt.Errorf("Error creating VM: ", err)
	}

	d.VMId = vmCreateTask.Entity.ID
	fmt.Println("VM was created with Id: ", d.VMId)

	vmStartTask, err := client.VMs.Start(d.VMId)
	if (err != nil) {
		return fmt.Errorf("Starting VM task not completed in time: ", err)
	}

	// Waiting for start VM task completion
	vmStartTask, err = client.Tasks.Wait(vmStartTask.ID)
	if (err != nil) {
		return fmt.Errorf("Error starting VM: ", err)
	}

	fmt.Println("VM is started.")

	ipAddress := ""
	numRetries := 0
	for numRetries < maxRetries {
		vmNetworksTask, err := client.VMs.GetNetworks(d.VMId)
		if (err != nil) {
			return fmt.Errorf("Error creating task for get VM networks: ", err)
		}

		// Waiting for get VM networks task completion
		vmNetworksTask, err = client.Tasks.Wait(vmNetworksTask.ID)
		if (err != nil) {
			return fmt.Errorf("Get VM networks taks not completed: ", err)
		}

		networkConnections := vmNetworksTask.ResourceProperties.(map[string]interface{})
		networks := networkConnections["networkConnections"].([]interface{})

		for _, nt := range networks {
			network := nt.(map[string]interface{})
			if val, ok := network["ipAddress"]; ok && val != nil {
				ipAddress = val.(string)
				break
			}
		}

		if ipAddress != "" {
			break;
		}
		numRetries++
		time.Sleep(1 * time.Second)
	}

	d.IPAddress = ipAddress
	fmt.Println("VM IP: ", d.VMId)

	return nil
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) GetState() (state.State, error) {
	client := d.getClient()
	vm, err := client.VMs.Get(d.VMId)
	if (err != nil) {
		return state.Error, fmt.Errorf("Error getting VM: ", err)
	}

	if (vm.State == "STOPPED") {
		return state.Stopped, nil
	}

	if (vm.State == "STARTED") {
		return state.Running, nil
	}

	if (vm.State == "SUSPENDED") {
		return state.Paused, nil
	}

	return state.Error, nil
}

func (d *Driver) Remove() error {
	// Stop the VM before attempting delete
	d.Stop()

	client := d.getClient()
	opTask, err := client.VMs.Delete(d.VMId)
	if (err != nil) {
		return fmt.Errorf("Error creating delete VM Task: ", err)
	}

	// Waiting for delete VM task completion
	opTask, err = client.Tasks.Wait(opTask.ID)
	if (err != nil) {
		return fmt.Errorf("Error deleting VM: ", err)
	}

	fmt.Println("VM was deleted.")

	return nil
}

func (d *Driver) Start() error {
	client := d.getClient()
	opTask, err := client.VMs.Start(d.VMId)
	if (err != nil) {
		return fmt.Errorf("Error creating start VM Task: ", err)
	}

	// Waiting for start VM task completion
	opTask, err = client.Tasks.Wait(opTask.ID)
	if (err != nil) {
		return fmt.Errorf("Error starting VM: ", err)
	}

	fmt.Println("VM was started.")

	return nil
}

func (d *Driver) Stop() error {
	client := d.getClient()
	opTask, err := client.VMs.Stop(d.VMId)
	if (err != nil) {
		return fmt.Errorf("Error creating stop VM task: ", err)
	}

	// Waiting for stop VM task completion
	opTask, err = client.Tasks.Wait(opTask.ID)
	if (err != nil) {
		return fmt.Errorf("Error stopping VM: ", err)
	}

	fmt.Println("VM was stopped.")

	return nil
}

func (d *Driver) Restart() error {
	client := d.getClient()
	opTask, err := client.VMs.Restart(d.VMId)
	if (err != nil) {
		return fmt.Errorf("Error creating restart VM task: ", err)
	}

	// Waiting for restart VM task completion
	opTask, err = client.Tasks.Wait(opTask.ID)
	if (err != nil) {
		return fmt.Errorf("Error restarting VM: ", err)
	}

	fmt.Println("VM was restarted.")

	return nil
}

func (d *Driver) Kill() error {
	return d.Stop()
}

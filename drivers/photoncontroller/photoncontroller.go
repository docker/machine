/*
 * Docker machine driver for VMware's Photon Controller: http://vmware.github.io/photon-controller
 */

package photoncontroller

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/state"
	"github.com/vmware/photon-controller-go-sdk/photon"
	cryptossh "golang.org/x/crypto/ssh"
)

const (
	driverName          = "photoncontroller"
	defaultProject      = "00000000-0000-0000-0000-000000000000"
	defaultVMFlavor     = "VMFlavor"
	defaultDiskFlavor   = "DiskFlavor"
	defaultImage        = "00000000-0000-0000-0000-000000000000"
	defaultDiskName     = "boot-disk"
	defaultBootDiskSize = 2
	defaultEndpoint     = "https://192.0.2.2"
	defaultAuthEndpoint = ""
	defaultSSHUser      = "docker"
	defaultSSHPort      = 22
)

type Driver struct {
	*drivers.BaseDriver
	Name            string
	Project         string
	VMFlavor        string
	DiskFlavor      string
	Image           string
	DiskName        string
	BootDiskSize    int
	VMId            string
	ISOPath         string
	SSHUserPassword string
	PhotonEndpoint  string
}

type NotLoadable struct {
	Name string
}

func (e NotLoadable) Error() string {
	return fmt.Sprintf("Driver %q not found. Do you have the plugin binary accessible in your PATH?", e.Name)
}

func NewDriver(hostName, storePath string) *Driver {
	return &Driver{
		Project:        defaultProject,
		VMFlavor:       defaultVMFlavor,
		DiskFlavor:     defaultDiskFlavor,
		Image:          defaultImage,
		DiskName:       defaultDiskName,
		BootDiskSize:   defaultBootDiskSize,
		PhotonEndpoint: defaultEndpoint,
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
			SSHPort:     defaultSSHPort,
		},
	}
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "photon-endpoint",
			Usage:  "Photon Controller endpoint in format: http://<ip address>:<port>",
			EnvVar: "PHOTON_ENDPOINT",
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
			Name:   "photon-image",
			Usage:  "Image Id",
			EnvVar: "PHOTON_IMAGE",
		},
		mcnflag.StringFlag{
			Name:   "photon-diskname",
			Usage:  "Disk name",
			Value:  defaultDiskName,
			EnvVar: "PHOTON_DISK_NAME",
		},
		mcnflag.IntFlag{
			Name:   "photon-bootdisksize",
			Usage:  "Boot disk size in GB",
			Value:  defaultBootDiskSize,
			EnvVar: "PHOTON_BOOT_DISK_SIZE",
		},
		mcnflag.StringFlag{
			Name:   "photon-iso-path",
			Usage:  "Path to ISO image with cloud-init data. Mutually exclusive with --photon-ssh-user-password",
			EnvVar: "PHOTON_ISO_PATH",
		},
		mcnflag.StringFlag{
			Name:   "photon-ssh-user-password",
			Usage:  "SSH User Password. Mutually exclusive with --photon-iso-path",
			EnvVar: "PHOTON_SSH_USER_PASSWORD",
		},
		mcnflag.StringFlag{
			Name:   "photon-ssh-keypath",
			Usage:  "SSH key path",
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
	return photon.NewClient(d.PhotonEndpoint, nil, nil)
}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}

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
	if err := mcnutils.WaitFor(d.RetrieveMachineIP); err != nil {
		return "", err
	}

	return d.IPAddress, nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHKeyPath() string {
	return d.SSHKeyPath
}

func (d *Driver) GetPublicSSHKeyPath() string {
	return d.SSHKeyPath + ".pub"
}

func (d *Driver) GetSSHPort() (int, error) {
	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	return d.SSHUser
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Project = flags.String("photon-project")
	d.VMFlavor = flags.String("photon-vmflavor")
	d.DiskFlavor = flags.String("photon-diskflavor")
	d.Image = flags.String("photon-image")
	d.DiskName = flags.String("photon-diskname")
	d.BootDiskSize = flags.Int("photon-bootdisksize")
	d.ISOPath = flags.String("photon-iso-path")
	d.SSHUserPassword = flags.String("photon-ssh-user-password")
	d.PhotonEndpoint = flags.String("photon-endpoint")
	d.SSHKeyPath = flags.String("photon-ssh-keypath")
	d.SSHUser = flags.String("photon-ssh-user")
	d.SSHPort = defaultSSHPort
	d.SetSwarmConfigFromFlags(flags)

	return nil
}

func (d *Driver) Create() error {
	vmSpec := &photon.VmCreateSpec{
		Name:          d.MachineName,
		Flavor:        d.VMFlavor,
		SourceImageID: d.Image,
		AttachedDisks: []photon.AttachedDisk{
			{
				CapacityGB: d.BootDiskSize,
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
	if err != nil {
		return fmt.Errorf("Error creating VM Task: %s", err.Error())
	}

	// Waiting for create VM task completion
	vmCreateTask, err = client.Tasks.Wait(vmCreateTask.ID)
	if err != nil {
		return fmt.Errorf("Error creating VM: %s", err.Error())
	}

	d.VMId = vmCreateTask.Entity.ID
	log.Info("VM was created with Id: ", d.VMId)

	if d.ISOPath != "" {
		// Creating task to attach ISO to VM. This is used to enable SSH access using public key defined in ISO.
		// Note: This relies on cloud-init ISO and that contains user-data.txt to configure it.
		if err = d.AttachISO(client); err != nil {
			return fmt.Errorf("Error attaching ISO to VM: %s", err.Error())
		}

		log.Info("ISO is attached to VM.")
	}

	d.Start()
	log.Info("VM is started.")
	log.Info("VM IP: ", d.IPAddress)

	return nil
}

func (d *Driver) PreCreateCheck() error {
	if d.PhotonEndpoint == "" {
		return fmt.Errorf("Photon controller endpoint was not specified. Use --photon-endpoint option to specify it.")
	}

	if d.Project == "" {
		return fmt.Errorf("Project Id was not provided. Use --photon-project option to specify it.")
	}

	if d.VMFlavor == "" {
		return fmt.Errorf("VM flavor name was not provided. Use --photon-vmflavor option to specify it")
	}

	if d.DiskFlavor == "" {
		return fmt.Errorf("Disk flavor name was not provided. Use --photon-diskflavor option to specify it.")
	}

	if d.Image == "" {
		return fmt.Errorf("Image Id was not provided. Use --photon-image option to specify it.")
	}

	if d.SSHKeyPath == "" {
		return fmt.Errorf("SSH key paht was not provided. Use --photon-ssh-keypath option to specify it")
	}

	if d.ISOPath == "" && d.SSHUserPassword == "" {
		return fmt.Errorf("Both SSH user password and ISO path were not provided. Provide either one of them using --photon-ssh-user-password or --photon-iso-path option.")
	}

	return nil
}

func (d *Driver) GetState() (state.State, error) {
	client := d.getClient()
	vm, err := client.VMs.Get(d.VMId)
	if err != nil {
		return state.Error, fmt.Errorf("Error getting VM: %s", err.Error())
	}

	if vm.State == "STOPPED" {
		return state.Stopped, nil
	}

	if vm.State == "STARTED" {
		return state.Running, nil
	}

	if vm.State == "SUSPENDED" {
		return state.Paused, nil
	}

	return state.Error, nil
}

func (d *Driver) Remove() error {
	// Stop the VM before attempting delete
	d.Stop()

	client := d.getClient()
	opTask, err := client.VMs.Delete(d.VMId)
	if err != nil {
		return fmt.Errorf("Error creating delete VM Task: %s", err.Error())
	}

	// Waiting for delete VM task completion
	opTask, err = client.Tasks.Wait(opTask.ID)
	if err != nil {
		return fmt.Errorf("Error deleting VM: %s", err.Error())
	}

	return nil
}

func (d *Driver) Start() error {
	client := d.getClient()
	opTask, err := client.VMs.Start(d.VMId)
	if err != nil {
		return fmt.Errorf("Error creating start VM Task: %s", err.Error())
	}

	// Waiting for start VM task completion
	opTask, err = client.Tasks.Wait(opTask.ID)
	if err != nil {
		return fmt.Errorf("Error starting VM: %s", err.Error())
	}

	d.waitForVM()

	// This is required for boot2docker image on startup as it has no persistent storage.
	if d.ISOPath == "" {
		log.Infof("Setting up SSH key for %s [%s]", d.MachineName, d.IPAddress)
		d.SetupSSHKey()
	}

	return nil
}

func (d *Driver) Stop() error {
	client := d.getClient()
	opTask, err := client.VMs.Stop(d.VMId)
	if err != nil {
		return fmt.Errorf("Error creating stop VM task: %s", err.Error())
	}

	// Waiting for stop VM task completion
	opTask, err = client.Tasks.Wait(opTask.ID)
	if err != nil {
		return fmt.Errorf("Error stopping VM: %s", err.Error())
	}

	return nil
}

func (d *Driver) Restart() error {
	client := d.getClient()
	opTask, err := client.VMs.Restart(d.VMId)
	if err != nil {
		return fmt.Errorf("Error creating restart VM task: %s", err.Error())
	}

	// Waiting for restart VM task completion
	opTask, err = client.Tasks.Wait(opTask.ID)
	if err != nil {
		return fmt.Errorf("Error restarting VM: %s", err.Error())
	}

	d.waitForVM()

	// This is required for boot2docker image on startup as it has no persistent storage.
	if d.ISOPath == "" {
		d.SetupSSHKey()
	}

	return nil
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) RetrieveMachineIP() bool {
	d.IPAddress = ""
	client := d.getClient()

	// Creating task to get VM networks
	vmNetworksTask, err := client.VMs.GetNetworks(d.VMId)
	if err != nil {
		log.Debug("Error creating task for get VM networks: %s", err.Error())
		return false
	}

	// Waiting for get VM networks task completion
	vmNetworksTask, err = client.Tasks.Wait(vmNetworksTask.ID)
	if err != nil {
		log.Debug("Get VM networks taks not completed: %s", err.Error())
		return false
	}

	// Retrieving IP address for the VM
	networkConnections := vmNetworksTask.ResourceProperties.(map[string]interface{})
	networks := networkConnections["networkConnections"].([]interface{})

	for _, nt := range networks {
		network := nt.(map[string]interface{})
		networkValue, ok := network["network"]
		if !ok || networkValue == nil || networkValue.(string) == "" {
			continue
		}

		ipAddressValue, ok := network["ipAddress"]
		if !ok || ipAddressValue == nil || ipAddressValue.(string) == "" {
			continue
		}

		d.IPAddress = ipAddressValue.(string)
		break
	}

	if d.IPAddress == "" {
		log.Debug("Fail to retrieve VM IP.")
		return false
	}

	return true
}

func (d *Driver) VMIsRunning() bool {
	st, err := d.GetState()
	if err != nil {
		log.Debug(err)
	}

	if st == state.Running {
		return true
	}
	return false
}

func (d *Driver) waitForVM() error {
	if err := mcnutils.WaitFor(d.VMIsRunning); err != nil {
		return err
	}

	if err := mcnutils.WaitFor(d.RetrieveMachineIP); err != nil {
		return err
	}

	return nil
}

func (d *Driver) AttachISO(client *photon.Client) error {
	file, err := os.Open(d.ISOPath)
	if err != nil {
		return err
	}

	attachISOTask, err := client.VMs.AttachISO(d.VMId, file, d.ISOPath)
	if err != nil {
		return fmt.Errorf("Attach ISO to VM task not completed in time: %s", err.Error())
	}

	// Waiting for attach ISO to VM task completion
	attachISOTask, err = client.Tasks.Wait(attachISOTask.ID)
	if err != nil {
		return fmt.Errorf("Error attaching ISO to VM: %s", err.Error())
	}

	return nil
}

func (d *Driver) SetupSSHKey() error {
	var keyfh *os.File
	var keycontent []byte
	var err error

	// create .ssh folder in users home
	if err = executeSSHCommand(fmt.Sprintf("mkdir -p /home/%s/.ssh", d.SSHUser), d); err != nil {
		return err
	}

	// read generated public ssh key
	if keyfh, err = os.Open(d.GetPublicSSHKeyPath()); err != nil {
		return err
	}
	defer keyfh.Close()

	if keycontent, err = ioutil.ReadAll(keyfh); err != nil {
		return err
	}

	// add public ssh key to authorized_keys
	if err := executeSSHCommand(fmt.Sprintf("echo '%s' >> /home/%s/.ssh/authorized_keys", string(keycontent), d.SSHUser), d); err != nil {
		return err
	}

	// make it secure
	if err := executeSSHCommand(fmt.Sprintf("chmod 600 /home/%s/.ssh/authorized_keys", d.SSHUser), d); err != nil {
		return err
	}

	return nil
}

// execute command over SSH with user / password authentication
func executeSSHCommand(command string, d *Driver) error {
	log.Debugf("Execute executeSSHCommand: %s", command)

	config := &cryptossh.ClientConfig{
		User: d.SSHUser,
		Auth: []cryptossh.AuthMethod{
			cryptossh.Password(d.SSHUserPassword),
		},
	}

	client, err := cryptossh.Dial("tcp", fmt.Sprintf("%s:%d", d.IPAddress, d.SSHPort), config)
	if err != nil {
		log.Debugf("Failed to dial:", err)
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		log.Debugf("Failed to create session: " + err.Error())
		return err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b

	if err := session.Run(command); err != nil {
		log.Debugf("Failed to run: " + err.Error())
		return err
	}
	log.Debugf("Stdout from executeSSHCommand: %s", b.String())

	return nil
}

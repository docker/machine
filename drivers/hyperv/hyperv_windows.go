package hyperv

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
)

const (
	isoFilename = "boot2docker-hyperv.iso"
)

type Driver struct {
	*drivers.BaseDriver
	boot2DockerURL string
	boot2DockerLoc string
	vSwitch        string
	diskImage      string
	diskSize       int
	memSize        int
}

func init() {
	drivers.Register("hyper-v", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "hyper-v-boot2docker-url",
			Usage: "Hyper-V URL of the boot2docker image. Defaults to the latest available version.",
		},
		cli.StringFlag{
			Name:  "hyper-v-boot2docker-location",
			Usage: "Hyper-V local boot2docker iso. Overrides URL.",
		},
		cli.StringFlag{
			Name:  "hyper-v-virtual-switch",
			Usage: "Hyper-V virtual switch name. Defaults to first found.",
		},
		cli.IntFlag{
			Name:  "hyper-v-disk-size",
			Usage: "Hyper-V disk size for host in MB.",
			Value: 20000,
		},
		cli.IntFlag{
			Name:  "hyper-v-memory",
			Usage: "Hyper-V memory size for host in MB.",
			Value: 1024,
		},
	}
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.boot2DockerURL = flags.String("hyper-v-boot2docker-url")
	d.boot2DockerLoc = flags.String("hyper-v-boot2docker-location")
	d.vSwitch = flags.String("hyper-v-virtual-switch")
	d.diskSize = flags.Int("hyper-v-disk-size")
	d.memSize = flags.Int("hyper-v-memory")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = "docker"
	d.SSHPort = 22
	return nil
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	inner := drivers.NewBaseDriver(machineName, storePath, caCert, privateKey)
	return &Driver{BaseDriver: inner}, nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "docker"
	}

	return d.SSHUser
}

func (d *Driver) DriverName() string {
	return "hyper-v"
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetState() (state.State, error) {

	command := []string{
		"(",
		"Get-VM",
		"-Name", d.MachineName,
		").state"}
	stdout, err := execute(command)
	if err != nil {
		return state.None, fmt.Errorf("Failed to find the VM status")
	}
	resp := parseStdout(stdout)

	if len(resp) < 1 {
		return state.None, nil
	}
	switch resp[0] {
	case "Running":
		return state.Running, nil
	case "Off":
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) Create() error {
	err := hypervAvailable()
	if err != nil {
		return err
	}

	d.setMachineNameIfNotSet()

	b2dutils := utils.NewB2dUtils("", "", isoFilename)
	if err := b2dutils.CopyIsoToMachineDir(d.boot2DockerURL, d.MachineName); err != nil {
		return err
	}

	log.Infof("Creating SSH key...")

	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating VM...")

	virtualSwitch, err := d.chooseVirtualSwitch()
	if err != nil {
		return err
	}

	if err := d.generateDiskImage(); err != nil {
		return err
	}

	command := []string{
		"New-VM",
		"-Name", d.MachineName,
		"-Path", fmt.Sprintf("'%s'", d.ResolveStorePath(".")),
		"-MemoryStartupBytes", fmt.Sprintf("%dMB", d.memSize)}
	_, err = execute(command)
	if err != nil {
		return err
	}

	command = []string{
		"Set-VMDvdDrive",
		"-VMName", d.MachineName,
		"-Path", fmt.Sprintf("'%s'", d.ResolveStorePath(isoFilename))}
	if _, err = execute(command); err != nil {
		return err
	}

	command = []string{
		"Add-VMHardDiskDrive",
		"-VMName", d.MachineName,
		"-Path", fmt.Sprintf("'%s'", d.diskImage)}
	if _, err = execute(command); err != nil {
		return err
	}

	command = []string{
		"Connect-VMNetworkAdapter",
		"-VMName", d.MachineName,
		"-SwitchName", fmt.Sprintf("'%s'", virtualSwitch)}
	if _, err = execute(command); err != nil {
		return err
	}

	log.Infof("Starting  VM...")
	if err := d.Start(); err != nil {
		return err
	}

	// use ssh to set keys
	sshClient, err := d.getLocalSSHClient()
	if err != nil {
		return err
	}

	// add pub key for user
	pubKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}

	if out, err := sshClient.Output(fmt.Sprintf(
		"mkdir -p /home/%s/.ssh",
		d.GetSSHUsername(),
	)); err != nil {
		log.Error(out)
		return err
	}

	if out, err := sshClient.Output(fmt.Sprintf(
		"printf '%%s' '%s' | tee /home/%s/.ssh/authorized_keys",
		string(pubKey),
		d.GetSSHUsername(),
	)); err != nil {
		log.Error(out)
		return err
	}

	return nil
}

func (d *Driver) chooseVirtualSwitch() (string, error) {
	if d.vSwitch != "" {
		return d.vSwitch, nil
	}
	command := []string{
		"@(Get-VMSwitch).Name"}
	stdout, err := execute(command)
	if err != nil {
		return "", err
	}
	switches := parseStdout(stdout)
	if len(switches) > 0 {
		log.Infof("Using switch %s", switches[0])
		return switches[0], nil
	}
	return "", fmt.Errorf("no vswitch found")
}

func (d *Driver) wait() error {
	log.Infof("Waiting for host to start...")
	for {
		ip, _ := d.GetIP()
		if ip != "" {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (d *Driver) Start() error {
	command := []string{
		"Start-VM",
		"-Name", d.MachineName}
	if _, err := execute(command); err != nil {
		return err
	}

	if err := d.wait(); err != nil {
		return err
	}

	if _, err := d.GetIP(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Stop() error {
	command := []string{
		"Stop-VM",
		"-Name", d.MachineName}
	if _, err := execute(command); err != nil {
		return err
	}
	for {
		s, err := d.GetState()
		if err != nil {
			return err
		}
		if s == state.Running {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	d.IPAddress = ""
	return nil
}

func (d *Driver) Remove() error {
	s, err := d.GetState()
	if err != nil {
		return err
	}
	if s == state.Running {
		if err := d.Kill(); err != nil {
			return err
		}
	}
	command := []string{
		"Remove-VM",
		"-Name", d.MachineName,
		"-Force"}
	if _, err = execute(command); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Restart() error {
	err := d.Stop()
	if err != nil {
		return err
	}

	return d.Start()
}

func (d *Driver) Kill() error {
	command := []string{
		"Stop-VM",
		"-Name", d.MachineName,
		"-TurnOff"}
	if _, err := execute(command); err != nil {
		return err
	}
	for {
		s, err := d.GetState()
		if err != nil {
			return err
		}
		if s == state.Running {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	d.IPAddress = ""
	return nil
}

func (d *Driver) setMachineNameIfNotSet() {
	if d.MachineName == "" {
		d.MachineName = fmt.Sprintf("docker-machine-unknown")
	}
}

func (d *Driver) GetIP() (string, error) {
	command := []string{
		"((",
		"Get-VM",
		"-Name", d.MachineName,
		").networkadapters[0]).ipaddresses[0]"}
	stdout, err := execute(command)
	if err != nil {
		return "", err
	}
	resp := parseStdout(stdout)
	if len(resp) < 1 {
		return "", fmt.Errorf("IP not found")
	}
	return resp[0], nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) generateDiskImage() error {
	d.diskImage = d.ResolveStorePath("disk.vhd")
	log.Infof("Creating VHD")
	command := []string{
		"New-VHD",
		"-Path", fmt.Sprintf("'%s'", d.diskImage),
		"-SizeBytes", fmt.Sprintf("%dMB", d.diskSize),
	}

	if _, err := execute(command); err != nil {
		return err
	}

	return nil
}

func (d *Driver) getLocalSSHClient() (ssh.Client, error) {
	ip, err := d.GetIP()
	if err != nil {
		return nil, err
	}

	sshAuth := &ssh.Auth{
		Passwords: []string{"docker"},
		Keys:      []string{d.GetSSHKeyPath()},
	}
	sshClient, err := ssh.NewNativeClient(d.GetSSHUsername(), ip, d.SSHPort, sshAuth)
	if err != nil {
		return nil, err
	}

	return sshClient, nil
}

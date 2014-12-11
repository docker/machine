package hyperv

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/utils"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

type Driver struct {
	storePath      string
	boot2DockerURL string
	boot2DockerLoc string
	vSwitch        string
	MachineName    string
	diskImage      string
	diskSize       int
	memSize        int
}

type CreateFlags struct {
	boot2DockerURL *string
	boot2DockerLoc *string
	vSwitch        *string
	diskSize       *int
	memSize        *int
}

func init() {
	drivers.Register("hyper-v", &drivers.RegisteredDriver{
		New:                 NewDriver,
		RegisterCreateFlags: RegisterCreateFlags,
	})
}

// RegisterCreateFlags registers the flags this driver adds to
// "docker hosts create"
func RegisterCreateFlags(cmd *flag.FlagSet) interface{} {
	createFlags := new(CreateFlags)
	createFlags.boot2DockerURL = cmd.String([]string{"-hyper-v-boot2docker-url"}, "", "The URL of the boot2docker image. Defaults to the latest available version.")
	createFlags.boot2DockerLoc = cmd.String([]string{"-hyper-v-boot2docker-location"}, "", "Local boot2docker iso. Overrides URL.")
	createFlags.vSwitch = cmd.String([]string{"-hyper-v-virtual-switch"}, "", "Name of virtual switch. Defaults to first found.")
	createFlags.diskSize = cmd.Int([]string{"-hyper-v-disk-size"}, 20000, "Size of disk for host in MB.")
	createFlags.memSize = cmd.Int([]string{"-hyper-v-memory"}, 1024, "Size of memory for host in MB.")
	return createFlags
}

func NewDriver(storePath string) (drivers.Driver, error) {
	return &Driver{storePath: storePath}, nil
}

func (d *Driver) DriverName() string {
	return "hyper-v"
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

func copyFile(inFile, outFile string) error {
	in, err := os.Open(inFile)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	err = out.Sync()
	return err
}

func (d *Driver) Create() error {
	err := hypervAvailable()
	if err != nil {
		return err
	}

	d.setMachineNameIfNotSet()

	var isoURL string

	if d.boot2DockerLoc == "" {
		if d.boot2DockerURL != "" {
			isoURL = d.boot2DockerURL
		} else {
			isoURL, err = getLatestReleaseURL()
			if err != nil {
				return err
			}
		}
		log.Infof("Downloading boot2docker...")

		if err := downloadISO(d.storePath, "boot2docker.iso", isoURL); err != nil {
			return err
		}
	} else {
		copyFile(d.boot2DockerLoc, filepath.Join(d.storePath, "boot2docker.iso"))
	}

	log.Infof("Creating SSH key...")

	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating VM...")

	virtualSwitch, err := d.chooseVirtualSwitch()
	if err != nil {
		return err
	}

	err = d.generateDiskImage()
	if err != nil {
		return err
	}

	command := []string{
		"New-VM",
		"-Name", d.MachineName,
		"-Path", fmt.Sprintf("'%s'", d.storePath),
		"-MemoryStartupBytes", fmt.Sprintf("%dMB", d.memSize)}
	_, err = execute(command)
	if err != nil {
		return err
	}

	command = []string{
		"Set-VMDvdDrive",
		"-VMName", d.MachineName,
		"-Path", fmt.Sprintf("'%s'", filepath.Join(d.storePath, "boot2docker.iso"))}
	_, err = execute(command)
	if err != nil {
		return err
	}

	command = []string{
		"Add-VMHardDiskDrive",
		"-VMName", d.MachineName,
		"-Path", fmt.Sprintf("'%s'", d.diskImage)}
	_, err = execute(command)
	if err != nil {
		return err
	}

	command = []string{
		"Connect-VMNetworkAdapter",
		"-VMName", d.MachineName,
		"-SwitchName", fmt.Sprintf("'%s'", virtualSwitch)}
	_, err = execute(command)
	if err != nil {
		return err
	}

	log.Infof("Starting  VM...")
	if err := d.Start(); err != nil {
		return err
	}

	log.Infof("Adding key to authorized-keys.d...")

	if err := drivers.AddPublicKeyToAuthorizedHosts(d, "/root/.docker/authorized-keys.d"); err != nil {
		return err
	}

	log.Infof("Restart docker...")
	cmd, err := d.GetSSHCommand("sudo /etc/init.d/docker restart")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
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

func (d *Driver) SetConfigFromFlags(flagsInterface interface{}) error {
	flags := flagsInterface.(*CreateFlags)
	d.boot2DockerURL = *flags.boot2DockerURL
	d.boot2DockerLoc = *flags.boot2DockerLoc
	d.vSwitch = *flags.vSwitch
	d.diskSize = *flags.diskSize
	d.memSize = *flags.memSize
	return nil
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
	log.Infof("Got IP, waiting for SSH")
	ip, _ := d.GetIP()
	return ssh.WaitForTCP(fmt.Sprintf("%s:22", ip))
}

func (d *Driver) Start() error {
	command := []string{
		"Start-VM",
		"-Name", d.MachineName}
	_, err := execute(command)
	if err != nil {
		return err
	}
	return d.wait()
}

func (d *Driver) Stop() error {
	command := []string{
		"Stop-VM",
		"-Name", d.MachineName}
	_, err := execute(command)
	if err != nil {
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
	_, err = execute(command)
	return err
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
	_, err := execute(command)
	if err != nil {
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
	return nil
}

func (d *Driver) setMachineNameIfNotSet() {
	if d.MachineName == "" {
		d.MachineName = fmt.Sprintf("docker-host-%s", utils.TruncateID(utils.GenerateRandomID()))
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

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	ip, _ := d.GetIP()
	return ssh.GetSSHCommand(ip, 22, "docker", d.sshKeyPath(), args...), nil
}

func (d *Driver) Upgrade() error {
	log.Infof("Stopping machine...")
	if err := d.Stop(); err != nil {
		return err
	}

	isoURL, err := getLatestReleaseURL()
	if err != nil {
		return err
	}

	log.Infof("Downloading boot2docker...")
	if err := downloadISO(d.storePath, "boot2docker.iso", isoURL); err != nil {
		return err
	}

	log.Infof("Starting machine...")
	if err := d.Start(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

// Get the latest boot2docker release tag name (e.g. "v0.6.0").
// FIXME: find or create some other way to get the "latest release" of boot2docker since the GitHub API has a pretty low rate limit on API requests
func getLatestReleaseURL() (string, error) {
	// HACK: boot2docker iso with ident auth built from here:
	// https://github.com/MSOpenTech/boot2docker/tree/ident-auth
	return "https://jlmstore.blob.core.windows.net/boot2docker/boot2docker-ident.iso", nil
	// 	rsp, err := http.Get("https://api.github.com/repos/boot2docker/boot2docker/releases")
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	defer rsp.Body.Close()

	// 	var t []struct {
	// 		TagName string `json:"tag_name"`
	// 	}
	// 	if err := json.NewDecoder(rsp.Body).Decode(&t); err != nil {
	// 		return "", err
	// 	}
	// 	if len(t) == 0 {
	// 		return "", fmt.Errorf("no releases found")
	// 	}

	// 	tag := t[0].TagName
	// 	url := fmt.Sprintf("https://github.com/boot2docker/boot2docker/releases/download/%s/boot2docker.iso", tag)
	// 	return url, nil
}

// Download boot2docker ISO image for the given tag and save it at dest.
func downloadISO(dir, file, url string) error {
	rsp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	// Download to a temp file first then rename it to avoid partial download.
	f, err := ioutil.TempFile(dir, file+".tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := io.Copy(f, rsp.Body); err != nil {
		// TODO: display download progress?
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(f.Name(), filepath.Join(dir, file)); err != nil {
		return err
	}
	return nil
}

func (d *Driver) generateDiskImage() error {
	// Create a small fixed vhd, put the tar in,
	// convert to dynamic, then resize

	d.diskImage = filepath.Join(d.storePath, "disk.vhd")
	fixed := filepath.Join(d.storePath, "fixed.vhd")
	log.Infof("Creating VHD")
	command := []string{
		"New-VHD",
		"-Path", fmt.Sprintf("'%s'", fixed),
		"-SizeBytes", "10MB",
		"-Fixed"}
	_, err := execute(command)
	if err != nil {
		return err
	}

	tarBuf, err := d.generateTar()
	if err != nil {
		return err
	}

	file, _ := os.OpenFile(fixed, os.O_WRONLY, 0644)
	defer file.Close()
	file.Seek(0, os.SEEK_SET)
	_, err = file.Write(tarBuf.Bytes())
	if err != nil {
		return err
	}
	file.Close()

	command = []string{
		"Convert-VHD",
		"-Path", fmt.Sprintf("'%s'", fixed),
		"-DestinationPath", fmt.Sprintf("'%s'", d.diskImage),
		"-VHDType", "Dynamic"}
	_, err = execute(command)
	if err != nil {
		return err
	}
	command = []string{
		"Resize-VHD",
		"-Path", fmt.Sprintf("'%s'", d.diskImage),
		"-SizeBytes", fmt.Sprintf("%dMB", d.diskSize)}
	_, err = execute(command)
	if err != nil {
		return err
	}

	return err
}

// Make a boot2docker VM disk image.
// See https://github.com/boot2docker/boot2docker/blob/master/rootfs/rootfs/etc/rc.d/automount
func (d *Driver) generateTar() (*bytes.Buffer, error) {
	magicString := "boot2docker, please format-me"

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	// magicString first so the automount script knows to format the disk
	file := &tar.Header{Name: magicString, Size: int64(len(magicString))}
	if err := tw.WriteHeader(file); err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(magicString)); err != nil {
		return nil, err
	}
	// .ssh/key.pub => authorized_keys
	file = &tar.Header{Name: ".ssh", Typeflag: tar.TypeDir, Mode: 0700}
	if err := tw.WriteHeader(file); err != nil {
		return nil, err
	}
	pubKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return nil, err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(pubKey)); err != nil {
		return nil, err
	}
	file = &tar.Header{Name: ".ssh/authorized_keys2", Size: int64(len(pubKey)), Mode: 0644}
	if err := tw.WriteHeader(file); err != nil {
		return nil, err
	}
	if _, err := tw.Write([]byte(pubKey)); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	return buf, nil
}

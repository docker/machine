package virtualbox

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/utils"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

type Driver struct {
	MachineName    string
	SSHPort        int
	Memory         int
	DiskSize       int
	Boot2DockerURL string
	storePath      string
}

type CreateFlags struct {
	Memory         *int
	DiskSize       *int
	Boot2DockerURL *string
}

func init() {
	drivers.Register("virtualbox", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// RegisterCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.IntFlag{
			Name:  "virtualbox-memory",
			Usage: "Size of memory for host in MB",
			Value: 1024,
		},
		cli.IntFlag{
			Name:  "virtualbox-disk-size",
			Usage: "Size of disk for host in MB",
			Value: 20000,
		},
		cli.StringFlag{
			EnvVar: "VIRTUALBOX_BOOT2DOCKER_URL",
			Name:   "virtualbox-boot2docker-url",
			Usage:  "The URL of the boot2docker image. Defaults to the latest available version",
			Value:  "",
		},
	}
}

func NewDriver(machineName string, storePath string) (drivers.Driver, error) {
	return &Driver{MachineName: machineName, storePath: storePath}, nil
}

func (d *Driver) DriverName() string {
	return "virtualbox"
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

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Memory = flags.Int("virtualbox-memory")
	d.DiskSize = flags.Int("virtualbox-disk-size")
	d.Boot2DockerURL = flags.String("virtualbox-boot2docker-url")
	return nil
}

func cpIso(src, dest string) error {
	buf, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(dest, buf, 0600); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Create() error {
	var (
		err    error
		isoURL string
	)

	// Check that VBoxManage exists and works
	if err = vbm(); err != nil {
		return err
	}

	d.SSHPort, err = getAvailableTCPPort()
	if err != nil {
		return err
	}

	if d.Boot2DockerURL != "" {
		isoURL = d.Boot2DockerURL
		log.Infof("Downloading boot2docker.iso from %s...", isoURL)
		if err := downloadISO(d.storePath, "boot2docker.iso", isoURL); err != nil {
			return err
		}
	} else {
		// HACK: Docker 1.4.1 boot2docker image with client/daemon auth
		isoURL = "https://ejhazlett.s3.amazonaws.com/public/boot2docker/machine-b2d-docker-1.4.1-identity.iso"

		// todo: check latest release URL, download if it's new
		// until then always use "latest"

		// isoURL, err = getLatestReleaseURL()
		// if err != nil {
		// 	return err
		// }

		// todo: use real constant for .docker
		rootPath := filepath.Join(drivers.GetHomeDir(), ".docker")
		imgPath := filepath.Join(rootPath, "images")
		commonIsoPath := filepath.Join(imgPath, "boot2docker.iso")
		if _, err := os.Stat(commonIsoPath); os.IsNotExist(err) {
			log.Infof("Downloading boot2docker.iso to %s...", commonIsoPath)

			// just in case boot2docker.iso has been manually deleted
			if _, err := os.Stat(imgPath); os.IsNotExist(err) {
				if err := os.Mkdir(imgPath, 0700); err != nil {
					return err
				}
			}

			if err := downloadISO(imgPath, "boot2docker.iso", isoURL); err != nil {
				return err
			}
		}

		isoDest := filepath.Join(d.storePath, "boot2docker.iso")
		if err := cpIso(commonIsoPath, isoDest); err != nil {
			return err
		}
	}

	log.Infof("Creating SSH key...")

	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating VirtualBox VM...")

	if err := d.generateDiskImage(d.DiskSize); err != nil {
		return err
	}

	if err := vbm("createvm",
		"--basefolder", d.storePath,
		"--name", d.MachineName,
		"--register"); err != nil {
		return err
	}

	cpus := uint(runtime.NumCPU())
	if cpus > 32 {
		cpus = 32
	}

	if err := vbm("modifyvm", d.MachineName,
		"--firmware", "bios",
		"--bioslogofadein", "off",
		"--bioslogofadeout", "off",
		"--natdnshostresolver1", "on",
		"--bioslogodisplaytime", "0",
		"--biosbootmenu", "disabled",

		"--ostype", "Linux26_64",
		"--cpus", fmt.Sprintf("%d", cpus),
		"--memory", fmt.Sprintf("%d", d.Memory),

		"--acpi", "on",
		"--ioapic", "on",
		"--rtcuseutc", "on",
		"--cpuhotplug", "off",
		"--pae", "on",
		"--synthcpu", "off",
		"--hpet", "on",
		"--hwvirtex", "on",
		"--nestedpaging", "on",
		"--largepages", "on",
		"--vtxvpid", "on",
		"--accelerate3d", "off",
		"--boot1", "dvd"); err != nil {
		return err
	}

	if err := vbm("modifyvm", d.MachineName,
		"--nic1", "nat",
		"--nictype1", "virtio",
		"--cableconnected1", "on"); err != nil {
		return err
	}

	if err := vbm("modifyvm", d.MachineName,
		"--natpf1", fmt.Sprintf("ssh,tcp,127.0.0.1,%d,,22", d.SSHPort)); err != nil {
		return err
	}

	hostOnlyNetwork, err := getOrCreateHostOnlyNetwork(
		net.ParseIP("192.168.99.1"),
		net.IPv4Mask(255, 255, 255, 0),
		net.ParseIP("192.168.99.2"),
		net.ParseIP("192.168.99.100"),
		net.ParseIP("192.168.99.254"))
	if err != nil {
		return err
	}
	if err := vbm("modifyvm", d.MachineName,
		"--nic2", "hostonly",
		"--nictype2", "virtio",
		"--hostonlyadapter2", hostOnlyNetwork.Name,
		"--cableconnected2", "on"); err != nil {
		return err
	}

	if err := vbm("storagectl", d.MachineName,
		"--name", "SATA",
		"--add", "sata",
		"--hostiocache", "on"); err != nil {
		return err
	}

	if err := vbm("storageattach", d.MachineName,
		"--storagectl", "SATA",
		"--port", "0",
		"--device", "0",
		"--type", "dvddrive",
		"--medium", filepath.Join(d.storePath, "boot2docker.iso")); err != nil {
		return err
	}

	if err := vbm("storageattach", d.MachineName,
		"--storagectl", "SATA",
		"--port", "1",
		"--device", "0",
		"--type", "hdd",
		"--medium", d.diskPath()); err != nil {
		return err
	}

	// let VBoxService do nice magic automounting (when it's used)
	if err := vbm("guestproperty", "set", d.MachineName, "/VirtualBox/GuestAdd/SharedFolders/MountPrefix", "/"); err != nil {
		return err
	}
	if err := vbm("guestproperty", "set", d.MachineName, "/VirtualBox/GuestAdd/SharedFolders/MountDir", "/"); err != nil {
		return err
	}

	var shareName, shareDir string // TODO configurable at some point
	switch runtime.GOOS {
	case "darwin":
		shareName = "Users"
		shareDir = "/Users"
		// TODO "linux" and "windows"
	}

	if shareDir != "" {
		if _, err := os.Stat(shareDir); err != nil && !os.IsNotExist(err) {
			return err
		} else if !os.IsNotExist(err) {
			if shareName == "" {
				// parts of the VBox internal code are buggy with share names that start with "/"
				shareName = strings.TrimLeft(shareDir, "/")
				// TODO do some basic Windows -> MSYS path conversion
				// ie, s!^([a-z]+):[/\\]+!\1/!; s!\\!/!g
			}

			// woo, shareDir exists!  let's carry on!
			if err := vbm("sharedfolder", "add", d.MachineName, "--name", shareName, "--hostpath", shareDir, "--automount"); err != nil {
				return err
			}

			// enable symlinks
			if err := vbm("setextradata", d.MachineName, "VBoxInternal2/SharedFoldersEnableSymlinksCreate/"+shareName, "1"); err != nil {
				return err
			}
		}
	}

	log.Infof("Starting VirtualBox VM...")

	if err := d.Start(); err != nil {
		return err
	}

	log.Debugf("Adding key to authorized-keys.d...")

	cmd, err := d.GetSSHCommand("sudo mkdir -p /var/lib/boot2docker/.docker && sudo chown -R docker /var/lib/boot2docker/.docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	if err := drivers.AddPublicKeyToAuthorizedHosts(d, "/var/lib/boot2docker/.docker/authorized-keys.d"); err != nil {
		return err
	}

	cmd, err = d.GetSSHCommand("if [ -e /var/run/docker.pid ]; then sudo /etc/init.d/docker stop; fi")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	// HACK: configure docker to use persisted auth
	cmd, err = d.GetSSHCommand("echo DOCKER_TLS=no | sudo tee -a /var/lib/boot2docker/profile")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	extraArgs := `EXTRA_ARGS='--auth=identity
        --auth-authorized-dir=/var/lib/boot2docker/.docker/authorized-keys.d
        --auth-known-hosts=/var/lib/boot2docker/.docker/known-hosts.json
        --identity=/var/lib/boot2docker/.docker/key.json
        -H tcp://0.0.0.0:2376'`
	sshCmd := fmt.Sprintf("echo \"%s\" | sudo tee -a /var/lib/boot2docker/profile", extraArgs)
	cmd, err = d.GetSSHCommand(sshCmd)
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = d.GetSSHCommand("sudo /etc/init.d/docker start")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = d.GetSSHCommand(fmt.Sprintf(
		"sudo hostname %s && echo \"%s\" | sudo tee /var/lib/boot2docker/etc/hostname",
		d.MachineName,
		d.MachineName,
	))

	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Start() error {
	if err := vbm("startvm", d.MachineName, "--type", "headless"); err != nil {
		return err
	}
	log.Infof("Waiting for VM to start...")
	return ssh.WaitForTCP(fmt.Sprintf("localhost:%d", d.SSHPort))
}

func (d *Driver) Stop() error {
	if err := vbm("controlvm", d.MachineName, "acpipowerbutton"); err != nil {
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
		if err == ErrMachineNotExist {
			log.Infof("machine does not exist, assuming it has been removed already")
			return nil
		}
		return err
	}
	if s == state.Running {
		if err := d.Kill(); err != nil {
			return err
		}
	}
	return vbm("unregistervm", "--delete", d.MachineName)
}

func (d *Driver) Restart() error {
	if err := d.Stop(); err != nil {
		return err
	}
	return d.Start()
}

func (d *Driver) Kill() error {
	return vbm("controlvm", d.MachineName, "poweroff")
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

func (d *Driver) GetState() (state.State, error) {
	stdout, stderr, err := vbmOutErr("showvminfo", d.MachineName,
		"--machinereadable")
	if err != nil {
		if reMachineNotFound.FindString(stderr) != "" {
			return state.Error, ErrMachineNotExist
		}
		return state.Error, err
	}
	re := regexp.MustCompile(`(?m)^VMState="(\w+)"$`)
	groups := re.FindStringSubmatch(stdout)
	if len(groups) < 1 {
		return state.None, nil
	}
	switch groups[1] {
	case "running":
		return state.Running, nil
	case "paused":
		return state.Paused, nil
	case "saved":
		return state.Saved, nil
	case "poweroff", "aborted":
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) setMachineNameIfNotSet() {
	if d.MachineName == "" {
		d.MachineName = fmt.Sprintf("docker-host-%s", utils.GenerateRandomID())
	}
}

func (d *Driver) GetIP() (string, error) {
	// DHCP is used to get the IP, so virtualbox hosts don't have IPs unless
	// they are running
	s, err := d.GetState()
	if err != nil {
		return "", err
	}
	if s != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}

	cmd, err := d.GetSSHCommand("ip addr show dev eth1")
	if err != nil {
		return "", err
	}

	// reset to nil as if using from Host Stdout is already set when using DEBUG
	cmd.Stdout = nil

	b, err := cmd.Output()
	if err != nil {
		return "", err
	}
	out := string(b)
	log.Debugf("SSH returned: %s\nEND SSH\n", out)
	// parse to find: inet 192.168.59.103/24 brd 192.168.59.255 scope global eth1
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		vals := strings.Split(strings.TrimSpace(line), " ")
		if len(vals) >= 2 && vals[0] == "inet" {
			return vals[1][:strings.Index(vals[1], "/")], nil
		}
	}

	return "", fmt.Errorf("No IP address found %s", out)
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return ssh.GetSSHCommand("localhost", d.SSHPort, "docker", d.sshKeyPath(), args...), nil
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

func (d *Driver) diskPath() string {
	return filepath.Join(d.storePath, "disk.vmdk")
}

// Get the latest boot2docker release tag name (e.g. "v0.6.0").
// FIXME: find or create some other way to get the "latest release" of boot2docker since the GitHub API has a pretty low rate limit on API requests
func getLatestReleaseURL() (string, error) {
	rsp, err := http.Get("https://api.github.com/repos/boot2docker/boot2docker/releases")
	if err != nil {
		return "", err
	}
	defer rsp.Body.Close()

	var t []struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(rsp.Body).Decode(&t); err != nil {
		return "", err
	}
	if len(t) == 0 {
		return "", fmt.Errorf("no releases found")
	}

	tag := t[0].TagName
	url := fmt.Sprintf("https://github.com/boot2docker/boot2docker/releases/download/%s/boot2docker.iso", tag)
	return url, nil
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
	// FIXME: why isn't this just using the vbm*() functions?
	cmd := exec.Command(vboxManageCmd, "convertfromraw", "stdin", dest,
		fmt.Sprintf("%d", sizeBytes), "--format", "VMDK")

	if os.Getenv("DEBUG") != "" {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	n, err := io.Copy(stdin, r)
	if err != nil {
		return err
	}

	// The total number of bytes written to stdin must match sizeBytes, or
	// VBoxManage.exe on Windows will fail. Fill remaining with zeros.
	if left := sizeBytes - n; left > 0 {
		if err := zeroFill(stdin, left); err != nil {
			return err
		}
	}

	// cmd won't exit until the stdin is closed.
	if err := stdin.Close(); err != nil {
		return err
	}

	return cmd.Wait()
}

// zeroFill writes n zero bytes into w.
func zeroFill(w io.Writer, n int64) error {
	const blocksize = 32 << 10
	zeros := make([]byte, blocksize)
	var k int
	var err error
	for n > 0 {
		if n > blocksize {
			k, err = w.Write(zeros)
		} else {
			k, err = w.Write(zeros[:n])
		}
		if err != nil {
			return err
		}
		n -= int64(k)
	}
	return nil
}

func getAvailableTCPPort() (int, error) {
	// FIXME: this has a race condition between finding an available port and
	// virtualbox using that port. Perhaps we should randomly pick an unused
	// port in a range not used by kernel for assigning ports
	ln, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer ln.Close()
	addr := ln.Addr().String()
	addrParts := strings.SplitN(addr, ":", 2)
	return strconv.Atoi(addrParts[1])
}

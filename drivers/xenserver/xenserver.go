package xenserver

import (
	"archive/tar"
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
	"github.com/nilshell/xmlrpc"
	"golang.org/x/net/context"

	xsclient "github.com/xenserver/go-xenserver-client"
)

const (
	isoOnlineURL        = "https://github.com/xenserver/boot2docker/releases/download/v1.7.0-xentools/boot2docker-v1.7.0-xentools.iso"
	isoFilename         = "boot2docker.iso"
	tarFilename         = "boot2docker.tar"
	osTemplateLabelName = "Other install media"
	B2D_USER            = "docker"
	B2D_PASS            = "tcuser"
)

type Driver struct {
	MachineName    string
	SSHUser        string
	SSHPort        int
	Server         string
	Username       string
	Password       string
	Boot2DockerURL string
	CPU            uint
	Memory         uint
	DiskSize       uint
	SR             string
	Network        string
	Host           string
	StorePath      string
	ISO            string
	TAR            string
	UploadTimeout  uint
	WaitTimeout    uint
	CaCertPath     string
	PrivateKeyPath string
	SwarmMaster    bool
	SwarmHost      string
	SwarmDiscovery string

	xenAPIClient *XenAPIClient
}

func init() {
	drivers.Register("xenserver", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "XENSERVER_SERVER",
			Name:   "xenserver-server",
			Usage:  "XenServer server hostname/IP for docker VM",
		},
		cli.StringFlag{
			EnvVar: "XENSERVER_USERNAME",
			Name:   "xenserver-username",
			Usage:  "XenServer username",
		},
		cli.StringFlag{
			EnvVar: "XENSERVER_PASSWORD",
			Name:   "xenserver-password",
			Usage:  "XenServer password",
		},
		cli.StringFlag{
			EnvVar: "XENSERVER_BOOT2DOCKER_URL",
			Name:   "xenserver-boot2docker-url",
			Usage:  "XenServer URL for boot2docker image",
			Value:  isoOnlineURL,
		},
		cli.IntFlag{
			EnvVar: "XENSERVER_VCPU_COUNT",
			Name:   "xenserver-vcpu-count",
			Usage:  "XenServer vCPU number for docker VM",
			Value:  1,
		},
		cli.IntFlag{
			EnvVar: "XENSERVER_MEMORY_SIZE",
			Name:   "xenserver-memory-size",
			Usage:  "XenServer size of memory for docker VM (in MB)",
			Value:  1024,
		},
		cli.IntFlag{
			EnvVar: "XENSERVER_DISK_SIZE",
			Name:   "xenserver-disk-size",
			Usage:  "XenServer size of disk for docker VM (in MB)",
			Value:  5120,
		},
		cli.StringFlag{
			EnvVar: "XENSERVER_SR_LABEL",
			Name:   "xenserver-sr-label",
			Usage:  "XenServer SR label where the docker VM will be attached",
		},
		cli.StringFlag{
			EnvVar: "XENSERVER_NETWORK_LABEL",
			Name:   "xenserver-network-label",
			Usage:  "XenServer network label where the docker VM will be attached",
		},
		cli.StringFlag{
			EnvVar: "XENSERVER_HOST_LABEL",
			Name:   "xenserver-host-label",
			Usage:  "XenServer host label where the docker VM will be run",
		},
		cli.IntFlag{
			EnvVar: "XENSERVER_UPLOAD_TIMEOUT",
			Name:   "xenserver-upload-timeout",
			Usage:  "XenServer upload VDI timeout(seconds)",
			Value:  5 * 60,
		},
		cli.IntFlag{
			EnvVar: "XENSERVER_WAIT_TIMEOUT",
			Name:   "xenserver-wait-timeout",
			Usage:  "XenServer wait VM start timeout(seconds)",
			Value:  30 * 60,
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	return &Driver{MachineName: machineName, StorePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey}, nil
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
	return filepath.Join(d.StorePath, "id_rsa")
}

func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = 22
	}

	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = B2D_USER
	}

	return d.SSHUser
}

func (d *Driver) DriverName() string {
	return "xenserver"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.SSHUser = B2D_USER
	d.SSHPort = 22
	d.Server = flags.String("xenserver-server")
	d.Username = flags.String("xenserver-username")
	d.Password = flags.String("xenserver-password")
	d.Boot2DockerURL = flags.String("xenserver-boot2docker-url")
	d.CPU = uint(flags.Int("xenserver-vcpu-count"))
	d.Memory = uint(flags.Int("xenserver-memory-size"))
	d.DiskSize = uint(flags.Int("xenserver-disk-size"))
	d.SR = flags.String("xenserver-sr-label")
	d.Network = flags.String("xenserver-network-label")
	d.Host = flags.String("xenserver-host-label")
	d.UploadTimeout = uint(flags.Int("xenserver-upload-timeout"))
	d.WaitTimeout = uint(flags.Int("xenserver-wait-timeout"))
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.ISO = filepath.Join(d.StorePath, isoFilename)
	d.TAR = filepath.Join(d.StorePath, tarFilename)

	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, _ := d.GetIP()
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	status, err := d.GetState()
	if err != nil {
		return "", err
	}
	if status != state.Running {
		return "", fmt.Errorf("Docker VM(%v) is not running", d.MachineName)
	}

	// Need Login first if it is a fresh session
	c, err := d.GetXenAPIClient()
	if err != nil {
		return "", err
	}

	// Get doker machine by label name
	vm, err := c.GetUniqueVMByNameLabel(d.MachineName)
	if err != nil {
		return "", err
	}

	// Get docker machine VM metrics
	metrics, err := vm.GetGuestMetrics()
	if err != nil {
		return "", err
	}

	// Get networks metrics
	networks, ok := metrics["networks"]
	if !ok {
		return "", fmt.Errorf("Docker VM(%v) get network metrics error: \"networks\" not presented", d.MachineName)
	}

	net, ok := networks.(xmlrpc.Struct)
	if !ok {
		return "", fmt.Errorf("Docker VM(%v) get network metrics error: \"networks\" is not a xmlrpc.Struct instance", d.MachineName)
	}

	for i := 0; i < len(net); i++ {
		ip, ok := net[fmt.Sprintf("%d/ip", i)]
		if ok && ip != "" {
			return ip.(string), nil
		}
	}

	return "", fmt.Errorf("Docker VM(%v) get ipaddr error", d.MachineName)
}

func (d *Driver) GetState() (state.State, error) {
	var err error

	// Need Login first if it is a fresh session
	c, err := d.GetXenAPIClient()
	if err != nil {
		return state.None, err
	}

	// Get docker machine VM by label name
	vm, err := c.GetUniqueVMByNameLabel(d.MachineName)
	if err != nil {
		return state.None, err
	}

	// Get doker machine VM power state
	powerState, err := vm.GetPowerState()
	if err != nil {
		return state.None, err
	}

	// https://github.com/xapi-project/xen-api/blob/cacbb2d3c5996efbf4cf117ab862439271a8cecd/ocaml/client_records/record_util.ml#L20
	switch strings.ToLower(powerState) {
	case "halted":
		return state.Stopped, nil
	case "paused":
		return state.Paused, nil
	case "running":
		return state.Running, nil
	case "suspended":
		return state.Paused, nil
	case "shutting down":
		return state.Stopping, nil
	case "migrating":
		return state.Running, nil
	}

	return state.None, nil
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Create() error {
	var err error

	d.setMachineNameIfNotSet()

	// Download boot2docker ISO from Internet
	var isoURL string
	b2dutils := utils.NewB2dUtils("", "")

	if d.Boot2DockerURL != "" {
		isoURL = d.Boot2DockerURL
	} else {
		isoURL, err = b2dutils.GetLatestBoot2DockerReleaseURL()
		if err != nil {
			log.Errorf("Unable to check for the latest release: %s", err)
			return err
		}
	}
	log.Infof("Downloading %s from %s...", isoFilename, isoURL)
	if err := b2dutils.DownloadISO(d.StorePath, isoFilename, isoURL); err != nil {
		return err
	}

	log.Infof("Logging into XenServer %s...", d.Server)
	c, err := d.GetXenAPIClient()
	if err != nil {
		return err
	}

	// Generate SSH Keys
	log.Infof("Creating SSH key...")

	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating ISO VDI...")

	// Get the SR
	var sr *xsclient.SR
	if d.SR == "" {
		sr, err = c.GetDefaultSR()
	} else {
		sr, err = c.GetUniqueSRByNameLabel(d.SR)
	}
	if err != nil {
		return err
	}

	isoFileInfo, err := os.Stat(d.ISO)
	if err != nil {
		return err
	}

	// Create the VDI
	isoVdi, err := sr.CreateVdi(isoFilename, isoFileInfo.Size())
	if err != nil {
		log.Errorf("Unable to create ISO VDI '%s': %v", isoFilename, err)
		return err
	}

	// Import the VDI
	if err = d.importVdi(isoVdi, d.ISO, time.Duration(d.UploadTimeout)*time.Second); err != nil {
		return err
	}

	isoVdiUuid, err := isoVdi.GetUuid()
	if err != nil {
		return err
	}

	log.Infof("Creating Disk VDI...")
	err = d.generateDiskImage()
	if err != nil {
		return err
	}

	// Create the VDI
	diskVdi, err := sr.CreateVdi("bootdocker disk", int64(d.DiskSize)*1024*1024)
	if err != nil {
		log.Errorf("Unable to create ISO VDI '%s': %v", "bootdocker disk", err)
		return err
	}
	if err = d.importVdi(diskVdi, d.TAR, time.Duration(d.UploadTimeout)*time.Second); err != nil {
		return err
	}

	diskVdiUuid, err := diskVdi.GetUuid()
	if err != nil {
		return err
	}

	log.Infof("Creating VM...")

	vm0, err := c.GetUniqueVMByNameLabel(osTemplateLabelName)
	if err != nil {
		return err
	}

	// Clone VM from VM template
	vm, err := vm0.Clone(fmt.Sprintf("__gui__%s", d.MachineName))
	if err != nil {
		return err
	}

	vmMacSeed, err := pseudoUuid()
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Errorf("Unable get local hostname")
	}

	otherConfig := map[string]string{
		"base_template_name":     osTemplateLabelName,
		"install-methods":        "cdrom,nfs,http,ftp",
		"linux_template":         "true",
		"mac_seed":               vmMacSeed,
		"docker-machine-creator": hostname,
	}
	err = vm.SetOtherConfig(otherConfig)
	if err != nil {
		return err
	}

	log.Infof("Provision VM...")
	err = vm.Provision()
	if err != nil {
		return err
	}

	// Set machine name
	err = vm.SetNameLabel(d.MachineName)
	if err != nil {
		return err
	}

	// Set vCPU number
	err = vm.SetVCPUsMax(d.CPU)
	if err != nil {
		return err
	}

	err = vm.SetVCPUsAtStartup(d.CPU)
	if err != nil {
		return err
	}

	platform_params := map[string]string{
		"acpi":             "1",
		"apic":             "true",
		"cores-per-socket": "1",
		"device_id":        "0001",
		"nx":               "true",
		"pae":              "true",
		"vga":              "std",
		"videoram":         "8",
		"viridian":         "false",
	}
	err = vm.SetPlatform(platform_params)
	if err != nil {
		return err
	}

	// Set machine memory size
	err = vm.SetStaticMemoryRange(uint64(d.Memory)*1024*1024, uint64(d.Memory)*1024*1024)
	if err != nil {
		return err
	}

	log.Infof("Add ISO VDI to VM...")
	diskVdi, err = c.GetVdiByUuid(isoVdiUuid)
	if err != nil {
		return err
	}

	err = vm.ConnectVdi(diskVdi, xsclient.Disk, "0")
	if err != nil {
		return err
	}

	log.Infof("Add Disk VDI to VM...")
	diskVdi, err = c.GetVdiByUuid(diskVdiUuid)
	if err != nil {
		return err
	}

	err = vm.ConnectVdi(diskVdi, xsclient.Disk, "1")
	if err != nil {
		return err
	}

	log.Infof("Add Network to VM...")
	var networks []*xsclient.Network
	if d.Network == "" {
		networks1, err := c.GetNetworks()
		if err != nil {
			return err
		}
		for _, network := range networks1 {
			otherConfig, err := network.GetOtherConfig()
			if err != nil {
				return err
			}
			isInternal, ok := otherConfig["is_host_internal_management_network"]
			if ok && isInternal == "true" {
				continue
			}
			automaitc, ok := otherConfig["automatic"]
			if ok && automaitc == "false" {
				continue
			}
			networks = append(networks, network)
		}
	} else {
		network, err := c.GetUniqueNetworkByNameLabel(d.Network)
		if err != nil {
			return err
		}
		networks = append(networks, network)
	}
	if len(networks) == 0 {
		return fmt.Errorf("Unable get available networks for %v", d.MachineName)
	}

	vifDevices, err := vm.GetAllowedVIFDevices()
	if err != nil {
		return err
	}
	if len(vifDevices) < len(networks) {
		log.Warnf("VM(%s) networks number is limited to %d.", d.MachineName, len(vifDevices))
		networks = networks[:len(vifDevices)]
	}

	for i, network := range networks {
		_, err = vm.ConnectNetwork(network, vifDevices[i])
		if err != nil {
			return err
		}
	}

	log.Infof("Starting VM...")
	if d.Host == "" {
		if err = vm.Start(false, false); err != nil {
			return err
		}
	} else {
		host, err := c.GetUniqueHostByNameLabel(d.Host)
		if err != nil {
			return err
		}
		if err = vm.StartOn(host, false, false); err != nil {
			return err
		}
	}

	if err := d.wait(time.Duration(d.WaitTimeout) * time.Second); err != nil {
		return err
	}

	log.Infof("VM Created.")

	return nil
}

func (d *Driver) wait(timeout time.Duration) (err error) {
	log.Infof("Waiting for VM to start...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	out := make(chan error, 1)
	go func(ctx context.Context, out chan<- error) {
		var ip string
		for {
			ip, _ = d.GetIP()
			if ip != "" {
				break
			}
			if t, ok := ctx.Deadline(); ok && time.Now().After(t) {
				out <- fmt.Errorf("Wait GetIP timed out")
				return
			}
			time.Sleep(1 * time.Second)
		}

		port, err := d.GetSSHPort()
		if err != nil {
			out <- err
			return
		}

		addr := fmt.Sprintf("%s:%d", ip, port)
		log.Infof("Got VM address(%v), Now waiting for SSH", addr)
		out <- ssh.WaitForTCP(addr)
	}(ctx, out)

	select {
	case err := <-out:
		return err
	case <-ctx.Done():
		return fmt.Errorf("Wait for VM to start timed out: %v", ctx.Err())
	}

	return nil
}

func (d *Driver) Start() error {
	var err error

	c, err := d.GetXenAPIClient()
	if err != nil {
		return err
	}

	vm, err := c.GetUniqueVMByNameLabel(d.MachineName)
	if err != nil {
		return err
	}

	if err = vm.Start(false, true); err != nil {
		return err
	}
	if err = d.wait(time.Duration(d.WaitTimeout) * time.Second); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Stop() error {
	var err error

	c, err := d.GetXenAPIClient()
	if err != nil {
		return err
	}

	vm, err := c.GetUniqueVMByNameLabel(d.MachineName)
	if err != nil {
		return err
	}

	if err = vm.CleanShutdown(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Remove() error {
	var err error

	c, err := d.GetXenAPIClient()
	if err != nil {
		return err
	}

	vm, err := c.GetUniqueVMByNameLabel(d.MachineName)
	if err != nil {
		return err
	}

	vmState, err := d.GetState()
	if err != nil {
		return err
	}

	switch vmState {
	case state.None:
		return fmt.Errorf("Unable get VM(%v) state.", d.MachineName)
	case state.Stopping:
		return fmt.Errorf("VM(%v) state is stopping, please wait.", d.MachineName)
	case state.Stopped:
		break
	case state.Paused:
		if err = vm.HardShutdown(); err != nil {
			return err
		}
	default:
		if err = vm.CleanShutdown(); err != nil {
			if err = vm.HardShutdown(); err != nil {
				return err
			}
		}
	}

	disks, err := vm.GetDisks()
	if err != nil {
		return err
	}

	for _, disk := range disks {
		if err = disk.Destroy(); err != nil {
			return err
		}
	}

	if err = vm.Destroy(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Restart() error {
	if err := d.Stop(); err != nil {
		return err
	}

	return d.Start()
}

func (d *Driver) Kill() error {
	var err error

	c, err := d.GetXenAPIClient()
	if err != nil {
		return err
	}

	vm, err := c.GetUniqueVMByNameLabel(d.MachineName)
	if err != nil {
		return err
	}

	if err = vm.HardShutdown(); err != nil {
		return err
	}
	return nil
}

func (d *Driver) setMachineNameIfNotSet() {
	if d.MachineName == "" {
		d.MachineName = fmt.Sprintf("docker-machine-unknown")
	}
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.StorePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

// Make a boot2docker VM disk image.
func (d *Driver) generateDiskImage() error {
	magicString := "boot2docker, please format-me"

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	log.Infof("Generate Disk Image...")
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
	if err := ioutil.WriteFile(d.TAR, buf.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

func (d *Driver) GetXenAPIClient() (*XenAPIClient, error) {
	if d.xenAPIClient == nil {
		c := NewXenAPIClient(d.Server, d.Username, d.Password)
		if err := c.Login(); err != nil {
			return nil, err
		}
		d.xenAPIClient = &c
	}
	return d.xenAPIClient, nil
}

func (d *Driver) importVdi(vdi *xsclient.VDI, filename string, timeout time.Duration) error {
	f, err := os.Open(filename)
	if err != nil {
		log.Errorf("Unable to open disk image '%s': %v", filename, err)
		return err
	}

	// Get file length
	fi, err := f.Stat()
	if err != nil {
		log.Errorf("Unable to stat disk image '%s': %v", filename, err)
		return err
	}

	task, err := vdi.Client.CreateTask()
	if err != nil {
		return fmt.Errorf("Unable to create task: %v", err)
	}
	urlStr := fmt.Sprintf("https://%s/import_raw_vdi?vdi=%s&session_id=%s&task_id=%s",
		vdi.Client.Host, vdi.Ref, vdi.Client.Session.(string), task.Ref)

	// Define a new http Transport which allows self-signed certs
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	req, err := http.NewRequest("PUT", urlStr, f)
	if err != nil {
		return err
	}
	req.ContentLength = fi.Size()

	resp, err := tr.RoundTrip(req)
	if err != nil {
		log.Errorf("Unable to upload VDI: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msg, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("xenserver reply %s: %v", resp.Status, string(msg))
		log.Errorf("Unable to upload VDI: %v", err)
		return err
	}

	log.Infof("Waiting Upload VDI task(%v) to complete...", task.Ref)
	if err = d.waitTask(task, timeout); err != nil {
		return err
	}
	return nil
}

func (d *Driver) waitTask(task *xsclient.Task, timeout time.Duration) error {
	timeout_at := time.Now().Add(timeout)

	for {
		status, err := task.GetStatus()
		if err != nil {
			return fmt.Errorf("Failed to get task status: %s", err.Error())
		}

		if status == xsclient.Success {
			log.Infof("Upload VDI task(%s) completed", task.Ref)
			break
		}

		switch status {
		case xsclient.Pending:
			progress, err := task.GetProgress()
			if err != nil {
				return fmt.Errorf("Failed to get progress: %s", err.Error())
			}
			log.Debugf("Upload %.0f%% complete", progress*100)
			log.Infof("Uploading...")
		case xsclient.Failure:
			errorInfo, err := task.GetErrorInfo()
			if err != nil {
				errorInfo = []string{fmt.Sprintf("furthermore, failed to get error info: %s", err.Error())}
			}
			return fmt.Errorf("Task failed: %s", errorInfo)
		case xsclient.Cancelling, xsclient.Cancelled:
			return fmt.Errorf("Task Cancelled")
		default:
			return fmt.Errorf("Unknown task status %v", status)
		}

		if time.Now().After(timeout_at) {
			return fmt.Errorf("Upload VDI task(%s) timed out", task.Ref)
		}

		time.Sleep(5 * time.Second)
	}

	return nil
}

func pseudoUuid() (string, error) {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", err
	}
	uuid := fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
	return uuid, nil
}

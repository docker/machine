/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwarevsphere

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/drivers/vmwarevsphere/errors"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
	cssh "golang.org/x/crypto/ssh"
)

const (
	DATASTORE_DIR      = "boot2docker-iso"
	B2D_ISO_NAME       = "boot2docker-vmw.iso"
	DEFAULT_CPU_NUMBER = 2
	dockerConfigDir    = "/var/lib/boot2docker"
	B2D_USER           = "docker"
	B2D_PASS           = "tcuser"
)

type Driver struct {
	MachineName    string
	SSHPort        int
	CPU            int
	Memory         int
	DiskSize       int
	Boot2DockerURL string
	IP             string
	Username       string
	Password       string
	Network        string
	Datastore      string
	Datacenter     string
	Pool           string
	HostIP         string
	StorePath      string
	ISO            string
	CaCertPath     string
	PrivateKeyPath string

	storePath string
}

type CreateFlags struct {
	CPU            *int
	Memory         *int
	DiskSize       *int
	Boot2DockerURL *string
	IP             *string
	Username       *string
	Password       *string
	Network        *string
	Datastore      *string
	Datacenter     *string
	Pool           *string
	HostIP         *string
}

func init() {
	drivers.Register("vmwarevsphere", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.IntFlag{
			EnvVar: "VSPHERE_CPU_COUNT",
			Name:   "vmwarevsphere-cpu-count",
			Usage:  "vSphere CPU number for docker VM",
			Value:  2,
		},
		cli.IntFlag{
			EnvVar: "VSPHERE_MEMORY_SIZE",
			Name:   "vmwarevsphere-memory-size",
			Usage:  "vSphere size of memory for docker VM (in MB)",
			Value:  2048,
		},
		cli.IntFlag{
			EnvVar: "VSPHERE_DISK_SIZE",
			Name:   "vmwarevsphere-disk-size",
			Usage:  "vSphere size of disk for docker VM (in MB)",
			Value:  20000,
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_BOOT2DOCKER_URL",
			Name:   "vmwarevsphere-boot2docker-url",
			Usage:  "vSphere URL for boot2docker image",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_VCENTER",
			Name:   "vmwarevsphere-vcenter",
			Usage:  "vSphere IP/hostname for vCenter",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_USERNAME",
			Name:   "vmwarevsphere-username",
			Usage:  "vSphere username",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_PASSWORD",
			Name:   "vmwarevsphere-password",
			Usage:  "vSphere password",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_NETWORK",
			Name:   "vmwarevsphere-network",
			Usage:  "vSphere network where the docker VM will be attached",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_DATASTORE",
			Name:   "vmwarevsphere-datastore",
			Usage:  "vSphere datastore for docker VM",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_DATACENTER",
			Name:   "vmwarevsphere-datacenter",
			Usage:  "vSphere datacenter for docker VM",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_POOL",
			Name:   "vmwarevsphere-pool",
			Usage:  "vSphere resource pool for docker VM",
		},
		cli.StringFlag{
			EnvVar: "VSPHERE_COMPUTE_IP",
			Name:   "vmwarevsphere-compute-ip",
			Usage:  "vSphere compute host IP where the docker VM will be instantiated",
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	return &Driver{MachineName: machineName, StorePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey}, nil
}

func (d *Driver) DriverName() string {
	return "vmwarevsphere"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.SSHPort = 22
	d.CPU = flags.Int("vmwarevsphere-cpu-count")
	d.Memory = flags.Int("vmwarevsphere-memory-size")
	d.DiskSize = flags.Int("vmwarevsphere-disk-size")
	d.Boot2DockerURL = flags.String("vmwarevsphere-boot2docker-url")
	d.IP = flags.String("vmwarevsphere-vcenter")
	d.Username = flags.String("vmwarevsphere-username")
	d.Password = flags.String("vmwarevsphere-password")
	d.Network = flags.String("vmwarevsphere-network")
	d.Datastore = flags.String("vmwarevsphere-datastore")
	d.Datacenter = flags.String("vmwarevsphere-datacenter")
	d.Pool = flags.String("vmwarevsphere-pool")
	d.HostIP = flags.String("vmwarevsphere-compute-ip")

	d.ISO = path.Join(d.storePath, "boot2docker.iso")

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
	if status != state.Running {
		return "", errors.NewInvalidStateError(d.MachineName)
	}
	vcConn := NewVcConn(d)
	rawIp, err := vcConn.VmFetchIp()
	if err != nil {
		return "", err
	}
	ip := strings.Trim(strings.Split(rawIp, "\n")[0], " ")
	return ip, nil
}

func (d *Driver) GetState() (state.State, error) {
	vcConn := NewVcConn(d)
	stdout, err := vcConn.VmInfo()
	if err != nil {
		return state.None, err
	}

	if strings.Contains(stdout, "poweredOn") {
		return state.Running, nil
	} else if strings.Contains(stdout, "poweredOff") {
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

// the current implementation does the following:
// 1. check whether the docker directory contains the boot2docker ISO
// 2. generate an SSH keypair
// 3. create a virtual machine with the boot2docker ISO mounted;
// 4. reconfigure the virtual machine network and disk size;
func (d *Driver) Create() error {
	if err := d.checkVsphereConfig(); err != nil {
		return err
	}

	var (
		isoURL string
		err    error
	)

	b2dutils := utils.NewB2dUtils("", "")

	if d.Boot2DockerURL != "" {
		isoURL = d.Boot2DockerURL
		log.Infof("Downloading boot2docker.iso from %s...", isoURL)
		if err := b2dutils.DownloadISO(d.storePath, "boot2docker.iso", isoURL); err != nil {
			return err

		}

	} else {
		// TODO: until vmw tools are merged into b2d master
		// we will use the iso from the vmware team
		//// todo: check latest release URL, download if it's new
		//// until then always use "latest"
		//isoURL, err = b2dutils.GetLatestBoot2DockerReleaseURL()
		//if err != nil {
		//	log.Warnf("Unable to check for the latest release: %s", err)

		//}

		isoURL := "https://github.com/cloudnativeapps/boot2docker/releases/download/v1.5.0-vmw/boot2docker-1.5.0-vmw.iso"

		rootPath := filepath.Join(utils.GetDockerDir())
		imgPath := filepath.Join(rootPath, "images")
		commonIsoPath := filepath.Join(imgPath, B2D_ISO_NAME)
		if _, err := os.Stat(commonIsoPath); os.IsNotExist(err) {
			log.Infof("Downloading boot2docker.iso to %s...", commonIsoPath)
			// just in case boot2docker.iso has been manually deleted
			if _, err := os.Stat(imgPath); os.IsNotExist(err) {
				if err := os.Mkdir(imgPath, 0700); err != nil {
					return err

				}

			}
			if err := b2dutils.DownloadISO(imgPath, B2D_ISO_NAME, isoURL); err != nil {
				return err

			}

		}
		isoDest := filepath.Join(d.storePath, B2D_ISO_NAME)
		if err := utils.CopyFile(commonIsoPath, isoDest); err != nil {
			return err

		}

	}

	log.Infof("Generating SSH Keypair...")
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return err
	}

	vcConn := NewVcConn(d)
	log.Infof("Uploading Boot2docker ISO ...")
	if err := vcConn.DatastoreMkdir(DATASTORE_DIR); err != nil {
		return err
	}

	if _, err := os.Stat(d.ISO); os.IsNotExist(err) {
		log.Errorf("Unable to find boot2docker ISO at %s", d.ISO)
		return errors.NewIncompleteVsphereConfigError(d.ISO)
	}

	if err := vcConn.DatastoreUpload(d.ISO); err != nil {
		return err
	}

	isoPath := fmt.Sprintf("%s/%s", DATASTORE_DIR, B2D_ISO_NAME)
	if err := vcConn.VmCreate(isoPath); err != nil {
		return err
	}

	log.Infof("Configuring the virtual machine %s... ", d.MachineName)
	if err := vcConn.VmDiskCreate(); err != nil {
		return err
	}

	if err := vcConn.VmAttachNetwork(); err != nil {
		return err
	}

	if err := d.Start(); err != nil {
		return err
	}

	log.Debugf("Setting hostname: %s", d.MachineName)
	cmd, err := d.GetSSHCommand(fmt.Sprintf(
		"echo \"127.0.0.1 %s\" | sudo tee -a /etc/hosts && sudo hostname %s && echo \"%s\" | sudo tee /etc/hostname",
		d.MachineName,
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
		vcConn := NewVcConn(d)
		err := vcConn.VmPowerOn()
		if err != nil {
			return err
		}
		// this step waits for the vm to start and fetch its ip address;
		// this guarantees that the opem-vmtools has started working...
		_, err = vcConn.VmFetchIp()
		if err != nil {
			return err
		}

		log.Infof("Configuring virtual machine %s... ", d.MachineName)

		key, err := ioutil.ReadFile(d.publicSSHKeyPath())
		if err != nil {
			return err
		}

		// so, vmrun above will not work without vmtools in b2d.  since getting stuff into TCL
		// is much more painful, we simply use the b2d password to get the initial public key
		// onto the machine.  from then on we use the pub key.  meh.
		sshConfig := &cssh.ClientConfig{
			User: B2D_USER,
			Auth: []cssh.AuthMethod{
				cssh.Password(B2D_PASS),
			},
		}

		ip, err := d.GetIP()
		if err != nil {
			return err
		}

		sshClient, err := cssh.Dial("tcp", fmt.Sprintf("%s:22", ip), sshConfig)
		if err != nil {
			return err
		}
		session, err := sshClient.NewSession()
		if err != nil {
			return err
		}
		if err := session.Run(fmt.Sprintf("mkdir /home/docker/.ssh && echo \"%s\" > /home/docker/.ssh/authorized_keys", string(key))); err != nil {
			return err
		}
		session.Close()

		return nil
	}
	return errors.NewInvalidStateError(d.MachineName)
}

func (d *Driver) Stop() error {
	vcConn := NewVcConn(d)
	err := vcConn.VmPowerOff()
	if err != nil {
		return err
	}
	return err
}

func (d *Driver) Remove() error {
	machineState, err := d.GetState()
	if err != nil {
		return err
	}
	if machineState == state.Running {
		if err = d.Stop(); err != nil {
			return fmt.Errorf("can't stop VM: %s", err)
		}
	}
	vcConn := NewVcConn(d)
	err = vcConn.VmDestroy()
	if err != nil {
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
	return d.Stop()
}

func (d *Driver) StartDocker() error {
	log.Debug("Starting Docker...")

	cmd, err := d.GetSSHCommand("sudo /etc/init.d/docker start")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) StopDocker() error {
	log.Debug("Stopping Docker...")

	cmd, err := d.GetSSHCommand("sudo /etc/init.d/docker stop")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) GetDockerConfigDir() string {
	return dockerConfigDir
}

func (d *Driver) Upgrade() error {
	return fmt.Errorf("upgrade is not supported for vsphere driver at this moment")
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	ip, err := d.GetIP()
	if err != nil {
		return nil, err
	}
	return ssh.GetSSHCommand(ip, d.SSHPort, "docker", d.sshKeyPath(), args...), nil
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.StorePath, "id_docker_host_vsphere")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

func (d *Driver) checkVsphereConfig() error {
	if d.IP == "" {
		return errors.NewIncompleteVsphereConfigError("vSphere IP")
	}
	if d.Username == "" {
		return errors.NewIncompleteVsphereConfigError("vSphere username")
	}
	if d.Password == "" {
		return errors.NewIncompleteVsphereConfigError("vSphere password")
	}
	if d.Network == "" {
		return errors.NewIncompleteVsphereConfigError("vSphere network")
	}
	if d.Datastore == "" {
		return errors.NewIncompleteVsphereConfigError("vSphere datastore")
	}
	if d.Datacenter == "" {
		return errors.NewIncompleteVsphereConfigError("vSphere datacenter")
	}
	return nil
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
	if err := os.Rename(f.Name(), path.Join(dir, file)); err != nil {
		return err
	}
	return nil
}

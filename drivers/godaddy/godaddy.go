package godaddy

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/docker/machine/drivers/godaddy/cloud"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
)

type Driver struct {
	*drivers.BaseDriver
	client            func() cloud.ClientWrapper
	createSSHKey      func() error
	APIBaseURL        string
	Image             string
	DataCenter        string
	Zone              string
	Spec              string
	SSHKey            string
	SSHKeyID          string
	APIKey            string
	ServerID          string
	UsingSharedSSHKey bool
	BootTimeout       int
}

const (
	defaultAPIBaseURL  = "https://api.godaddy.com"
	defaultImage       = "ubuntu-1604"
	defaultDataCenter  = "phx"
	defaultSpec        = "tiny"
	defaultSSHUser     = "machine"
	defaultBootTimeout = 120
)

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			Name:   "godaddy-api-key",
			Usage:  "GoDaddy API Key",
			EnvVar: "GODADDY_API_KEY",
		},
		mcnflag.StringFlag{
			Name:   "godaddy-base-api-url",
			Usage:  "GoDaddy API URL",
			Value:  defaultAPIBaseURL,
			EnvVar: "GODADDY_API_URL",
		},
		mcnflag.StringFlag{
			Name:   "godaddy-image",
			Usage:  "GoDaddy Image",
			Value:  defaultImage,
			EnvVar: "GODADDY_IMAGE",
		},
		mcnflag.StringFlag{
			Name:   "godaddy-data-center",
			Usage:  "GoDaddy Data Center",
			Value:  defaultDataCenter,
			EnvVar: "GODADDY_DATA_CENTER",
		},
		mcnflag.StringFlag{
			Name:   "godaddy-zone",
			Usage:  "GoDaddy Zone",
			EnvVar: "GODADDY_ZONE",
		},
		mcnflag.StringFlag{
			Name:   "godaddy-spec",
			Usage:  "GoDaddy Spec",
			Value:  defaultSpec,
			EnvVar: "GODADDY_SPEC",
		},
		mcnflag.StringFlag{
			Name:   "godaddy-ssh-user",
			Usage:  "Name of the user to be used for SSH",
			Value:  defaultSSHUser,
			EnvVar: "GODADDY_SSH_USER",
		},
		mcnflag.StringFlag{
			Name:   "godaddy-ssh-key-id",
			Usage:  "Id of the GoDaddy SSH Key to associate with this new server",
			EnvVar: "GODADDY_SSH_KEY_ID",
		},
		mcnflag.StringFlag{
			Name:   "godaddy-ssh-key",
			Usage:  "Private keyfile to use for SSH (absolute path)",
			EnvVar: "GODADDY_SSH_KEY",
		},
		mcnflag.IntFlag{
			Name:   "godaddy-boot-timeout",
			Usage:  "Amount of time (in seconds) to wait for initial boot",
			Value:  defaultBootTimeout,
			EnvVar: "GODADDY_BOOT_TIMEOUT",
		},
	}
}

func NewDriver(hostName, storePath string) *Driver {
	driver := &Driver{
		BaseDriver: &drivers.BaseDriver{
			SSHUser:     defaultSSHUser,
			MachineName: hostName,
			StorePath:   storePath,
		},
	}
	driver.client = driver.buildClient
	driver.createSSHKey = func() error { return createSSHKey(driver) }
	return driver
}

func (d *Driver) buildClient() cloud.ClientWrapper {
	return cloud.NewClient(d.APIBaseURL, d.APIKey)
}

func (d *Driver) DriverName() string {
	return "godaddy"
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHUsername() string {
	return d.SSHUser
}

func (d *Driver) GetSSHKeyPath() string {
	if d.SSHKeyPath == "" && d.SSHKeyID == "" {
		d.SSHKeyPath = d.ResolveStorePath("id_rsa")
	}
	return d.SSHKeyPath
}

func (d *Driver) PreCreateCheck() error {
	if d.SSHKey != "" {
		if _, err := os.Stat(d.SSHKey); os.IsNotExist(err) {
			return fmt.Errorf("SSH key does not exist: %q", d.SSHKey)
		}
	}
	if d.SSHKeyID != "" && d.SSHKey == "" {
		return fmt.Errorf("Please provide the public key associated with %q via --godaddy-ssh-key", d.SSHKeyID)
	}
	return nil
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.APIBaseURL = flags.String("godaddy-base-api-url")
	d.SSHUser = flags.String("godaddy-ssh-user")
	d.Image = flags.String("godaddy-image")
	d.DataCenter = flags.String("godaddy-data-center")
	d.Zone = flags.String("godaddy-zone")
	d.Spec = flags.String("godaddy-spec")
	d.SSHKey = flags.String("godaddy-ssh-key")
	d.SSHKeyID = flags.String("godaddy-ssh-key-id")
	d.BootTimeout = flags.Int("godaddy-boot-timeout")
	if d.APIKey = flags.String("godaddy-api-key"); d.APIKey == "" {
		return errors.New("godaddy-api-key is required")
	}

	return nil
}

func (d *Driver) Create() error {
	// create or copy existing ssh key to store
	if err := d.createSSHKey(); err != nil {
		return err
	}

	servers := d.client().Servers
	args := cloud.ServerCreate{
		Hostname:     d.MachineName,
		Username:     d.SSHUser,
		Image:        d.Image,
		Spec:         d.Spec,
		DataCenterId: d.DataCenter,
		ZoneId:       d.Zone,
		SshKeyId:     d.SSHKeyID,
	}

	server, err := servers().AddServer(args)
	if err != nil {
		return err
	}
	d.ServerID = server.ServerId
	log.Debug("New server create accepted. Id:", d.ServerID)

	d.waitForIP()
	if err := d.waitForCloudInit(); err != nil {
		log.Info("Failed to verify boot prior to timeout. Continuing...", err)
	}

	return nil
}

func createSSHKey(d *Driver) error {
	if d.SSHKey == "" {
		log.Info("No SSH key specified. Generating new key...")
		if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
			return err
		}
		d.SSHKey = d.GetSSHKeyPath()
		log.Info("New key:", d.SSHKey)
	} else {
		log.Info("Importing SSH key...")
		d.SSHKeyPath = d.ResolveStorePath(path.Base(d.SSHKey))
		if err := copySSHKey(d.SSHKey, d.SSHKeyPath); err != nil {
			return err
		}
		if err := copySSHKey(d.SSHKey+".pub", d.SSHKeyPath+".pub"); err != nil {
			log.Infof("Couldn't copy SSH public key : %s", err)
			return err
		}
	}

	if d.SSHKeyID == "" {
		key, err := d.uploadSSHKey()
		if err != nil {
			return err
		}
		d.SSHKeyID = key.SshKeyId
	} else {
		d.UsingSharedSSHKey = true
	}

	return nil
}

func copySSHKey(src, dst string) error {
	if err := mcnutils.CopyFile(src, dst); err != nil {
		return fmt.Errorf("unable to copy ssh key: %s", err)
	}

	if err := os.Chmod(dst, 0600); err != nil {
		return fmt.Errorf("unable to set permissions on the ssh key: %s", err)
	}

	return nil
}

func (d *Driver) uploadSSHKey() (*cloud.SSHKey, error) {
	log.Info("Creating GoDaddy Cloud Servers SSH Key...")
	publicKey, err := ioutil.ReadFile(d.GetSSHKeyPath() + ".pub")
	if err != nil {
		return nil, err
	}

	args := cloud.SSHKeyCreate{
		Name: d.MachineName,
		Key:  string(publicKey),
	}
	newKey, err := d.client().SSHKeys().AddSSHKey(args)
	if err != nil {
		return nil, err
	}

	return &newKey, nil
}

func (d *Driver) waitForIP() error {
	log.Info("Waiting for public IP")
	return mcnutils.WaitFor(func() bool {
		server, err := d.client().Servers().GetServerById(d.ServerID)
		if err != nil {
			log.Debug(err)
			return false
		}

		d.IPAddress = server.PublicIp
		if d.IPAddress != "" {
			log.Info(d.IPAddress)
			return true
		}
		return false
	})
}

func (d *Driver) waitForCloudInit() error {
	if d.BootTimeout == 0 {
		return nil
	}

	log.Info("Waiting for initial boot")
	return mcnutils.WaitForSpecific(func() bool {
		_, err := drivers.RunSSHCommandFromDriver(d, "test -f /var/lib/cloud/instance/boot-finished")
		if err == nil {
			return true
		}

		log.Debug("boot not finished:", err)
		return false
	}, (d.BootTimeout / 10), 10*time.Second)
}

func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, strconv.Itoa(engine.DefaultPort))), nil
}

func (d *Driver) GetState() (state.State, error) {
	log.Debug("Get status for GoDaddy Cloud Server:", d.ServerID)
	s, err := d.client().Servers().GetServerById(d.ServerID)
	if err != nil {
		return state.None, err
	}

	log.Debug("mapping server status:", s.Status)
	switch s.Status {
	case "NEW", "BUILDING", "CONFIGURING_NETWORK", "VERIFYING", "STARTING":
		return state.Starting, nil
	case "RUNNING":
		return state.Running, nil
	case "DESTROYED":
		return state.None, nil
	case "STOPPING", "DESTROYING":
		return state.Stopping, nil
	case "STOPPED":
		return state.Stopped, nil
	case "ERROR":
		return state.Error, nil
	}

	return state.None, nil
}

func (d *Driver) Start() error {
	if _, err := d.client().Servers().StartServer(d.ServerID); err != nil {
		return err
	}
	err := mcnutils.WaitFor(drivers.MachineInState(d, state.Running))
	return err
}

func (d *Driver) Stop() error {
	if _, err := drivers.RunSSHCommandFromDriver(d, "sudo shutdown -h now"); err != nil {
		if !strings.Contains(err.Error(), "closed by remote host") {
			return err
		}
	}
	return nil
}

func (d *Driver) Restart() error {
	if _, err := drivers.RunSSHCommandFromDriver(d, "sudo shutdown -r now"); err != nil {
		if !strings.Contains(err.Error(), "closed by remote host") {
			return err
		}
	}
	return nil
}

func (d *Driver) Kill() error {
	if _, err := d.client().Servers().StopServer(d.ServerID); err != nil {
		return err
	}
	err := mcnutils.WaitFor(drivers.MachineInState(d, state.Stopped))
	return err
}

func (d *Driver) Remove() error {
	if d.ServerID == "" {
		return nil
	}
	client := d.client()
	if d.SSHKeyID != "" && !d.UsingSharedSSHKey {
		if err := client.SSHKeys().DeleteSSHKey(d.SSHKeyID); err != nil {
			log.Infof("Error removing SSHKey %s. %s", d.SSHKeyID, err)
		}
	}

	if _, err := client.Servers().DestroyServer(d.ServerID); err != nil {
		return err
	}
	return nil
}

package vultr

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	vultr "github.com/JamesClonk/vultr/lib"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

const (
	dockerConfigDir = "/etc/docker"
)

type Driver struct {
	APIKey         string
	MachineID      string
	MachineName    string
	IPAddress      string
	OSID           int
	RegionID       int
	PlanID         int
	SSHKeyID       string
	CaCertPath     string
	PrivateKeyPath string
	DriverKeyPath  string
	storePath      string
}

func init() {
	drivers.Register("vultr", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "VULTR_API_KEY",
			Name:   "vultr-api-key",
			Usage:  "Vultr API key",
		},
		cli.IntFlag{
			EnvVar: "VULTR_OS",
			Name:   "vultr-os-id",
			Usage:  "Vultr operating system ID (OSID). Default: 160 (Ubuntu 14.04 x64)",
			Value:  160,
		},
		cli.IntFlag{
			EnvVar: "VULTR_REGION",
			Name:   "vultr-region-id",
			Usage:  "Vultr region ID (DCID). Default: 1 (New Jersey)",
			Value:  1,
		},
		cli.IntFlag{
			EnvVar: "VULTR_PLAN",
			Name:   "vultr-plan-id",
			Usage:  "Vultr plan ID (VPSPLANID). Default: 29 (768 MB RAM, 15 GB SSD, 1.00 TB BW)",
			Value:  29,
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	return &Driver{MachineName: machineName, storePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey}, nil
}

func (d *Driver) DriverName() string {
	return "vultr"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.APIKey = flags.String("vultr-api-key")
	d.OSID = flags.Int("vultr-os-id")
	d.RegionID = flags.Int("vultr-region-id")
	d.PlanID = flags.Int("vultr-plan-id")

	if d.APIKey == "" {
		return fmt.Errorf("vultr driver requires the --vultr-api-key option")
	}
	return nil
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Create() error {
	log.Infof("Creating SSH key...")

	key, err := d.createSSHKey()
	if err != nil {
		return err
	}
	d.SSHKeyID = key.ID

	log.Infof("Creating Vultr virtual machine...")
	client := d.getClient()
	machine, err := client.CreateServer(
		d.MachineName,
		d.RegionID,
		d.PlanID,
		d.OSID,
		&vultr.ServerOptions{
			SSHKey: d.SSHKeyID,
		})
	if err != nil {
		return err
	}
	d.MachineID = machine.ID

	for {
		machine, err = client.GetServer(d.MachineID)
		if err != nil {
			return err
		}
		d.IPAddress = machine.MainIP

		if d.IPAddress != "" && d.IPAddress != "0" {
			break
		}

		time.Sleep(1 * time.Second)
	}

	log.Debugf("Created virtual machine: SUBID %s, IP address %s",
		machine.ID,
		d.IPAddress)

	log.Infof("Waiting for SSH...")
	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", d.IPAddress, 22)); err != nil {
		return err
	}

	// wait for dpkg lock to appear
	time.Sleep(15 * time.Second)

	log.Infof("Waiting for dpkg unlock...")
	for {
		cmd, err := d.GetSSHCommand(`lsof /var/lib/dpkg/lock >/dev/null 2>&1; [ $? = 0 ] && echo "locked"; echo ""`)
		if err != nil {
			return err
		}

		var out bytes.Buffer
		go io.Copy(cmd.Stdout, &out)

		if err := cmd.Run(); err != nil {
			return err
		}

		if !strings.Contains(out.String(), "locked") {
			break
		}

		log.Debugf("dpkg is still locked..")
		time.Sleep(15 * time.Second)
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

	log.Debugf("Installing Docker")

	cmd, err = d.GetSSHCommand("if [ ! -e /usr/bin/docker ]; then curl -sL https://get.docker.com | sh -; fi")
	if err != nil {
		return err

	}
	if err := cmd.Run(); err != nil {
		return err

	}

	return nil
}

func (d *Driver) createSSHKey() (*vultr.SSHKey, error) {
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return nil, err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return nil, err
	}

	key, err := d.getClient().CreateSSHKey(d.MachineName, string(publicKey))
	if err != nil {
		return &key, err
	}
	return &key, nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	if d.IPAddress == "" || d.IPAddress == "0" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	machine, err := d.getClient().GetServer(d.MachineID)
	if err != nil {
		return state.Error, err
	}
	switch machine.Status {
	case "pending":
		return state.Starting, nil
	case "active":
		switch machine.PowerStatus {
		case "running":
			return state.Running, nil
		case "stopped":
			return state.Stopped, nil
		}
	}
	return state.None, nil
}

func (d *Driver) Start() error {
	return d.getClient().StartServer(d.MachineID)
}

func (d *Driver) Stop() error {
	return d.getClient().HaltServer(d.MachineID)
}

func (d *Driver) Remove() error {
	client := d.getClient()
	if err := client.DeleteSSHKey(d.SSHKeyID); err != nil {
		return err
	}
	if err := client.DeleteServer(d.MachineID); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Restart() error {
	return d.getClient().RebootServer(d.MachineID)
}

func (d *Driver) Kill() error {
	return d.getClient().HaltServer(d.MachineID)
}

func (d *Driver) StartDocker() error {
	log.Debug("Starting Docker...")

	cmd, err := d.GetSSHCommand("sudo service docker start")
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

	cmd, err := d.GetSSHCommand("sudo service docker stop")
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
	log.Debugf("Upgrading Docker")

	cmd, err := d.GetSSHCommand("sudo apt-get update && sudo apt-get install --upgrade lxc-docker")
	if err != nil {
		return err

	}
	if err := cmd.Run(); err != nil {
		return err

	}

	return cmd.Run()
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return ssh.GetSSHCommand(d.IPAddress, 22, "root", d.sshKeyPath(), args...), nil
}

func (d *Driver) getClient() *vultr.Client {
	return vultr.NewClient(d.APIKey, nil)
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

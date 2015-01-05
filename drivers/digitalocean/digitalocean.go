package digitalocean

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"code.google.com/p/goauth2/oauth"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/digitalocean/godo"
	// "github.com/docker/docker/utils"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

type Driver struct {
	AccessToken string
	DropletID   int
	Image       string
	IPAddress   string
	MachineName string
	Region      string
	SSHKeyID    int
	Size        string
	storePath   string
}

func init() {
	drivers.Register("digitalocean", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "DIGITALOCEAN_ACCESS_TOKEN",
			Name:   "digitalocean-access-token",
			Usage:  "Digital Ocean access token",
		},
		cli.StringFlag{
			EnvVar: "DIGITALOCEAN_IMAGE",
			Name:   "digitalocean-image",
			Usage:  "Digital Ocean Image",
			Value:  "docker",
		},
		cli.StringFlag{
			EnvVar: "DIGITALOCEAN_REGION",
			Name:   "digitalocean-region",
			Usage:  "Digital Ocean region",
			Value:  "nyc3",
		},
		cli.StringFlag{
			EnvVar: "DIGITALOCEAN_SIZE",
			Name:   "digitalocean-size",
			Usage:  "Digital Ocean size",
			Value:  "512mb",
		},
	}
}

func NewDriver(machineName string, storePath string) (drivers.Driver, error) {
	return &Driver{MachineName: machineName, storePath: storePath}, nil
}

func (d *Driver) DriverName() string {
	return "digitalocean"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AccessToken = flags.String("digitalocean-access-token")
	d.Image = flags.String("digitalocean-image")
	d.Region = flags.String("digitalocean-region")
	d.Size = flags.String("digitalocean-size")

	if d.AccessToken == "" {
		return fmt.Errorf("digitalocean driver requires the --digitalocean-access-token option")
	}

	return nil
}

func (d *Driver) Create() error {
	log.Infof("Creating SSH key...")

	key, err := d.createSSHKey()
	if err != nil {
		return err
	}

	d.SSHKeyID = key.ID

	log.Infof("Creating Digital Ocean droplet...")

	client := d.getClient()

	createRequest := &godo.DropletCreateRequest{
		Image:   d.Image,
		Name:    d.MachineName,
		Region:  d.Region,
		Size:    d.Size,
		SSHKeys: []interface{}{d.SSHKeyID},
	}

	newDroplet, _, err := client.Droplets.Create(createRequest)
	if err != nil {
		return err
	}

	d.DropletID = newDroplet.Droplet.ID

	for {
		newDroplet, _, err = client.Droplets.Get(d.DropletID)
		if err != nil {
			return err
		}
		for _, network := range newDroplet.Droplet.Networks.V4 {
			if network.Type == "public" {
				d.IPAddress = network.IPAddress
			}
		}

		if d.IPAddress != "" {
			break
		}

		time.Sleep(1 * time.Second)
	}

	log.Debugf("Created droplet ID %d, IP address %s",
		newDroplet.Droplet.ID,
		d.IPAddress)

	log.Infof("Waiting for SSH...")

	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", d.IPAddress, 22)); err != nil {
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

	log.Debugf("HACK: Downloading version of Docker with identity auth...")

	cmd, err = d.GetSSHCommand("stop docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = d.GetSSHCommand("curl -sS https://ehazlett.s3.amazonaws.com/public/docker/linux/docker-1.4.1-136b351e-identity > /usr/bin/docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("Updating /etc/default/docker to use identity auth...")

	cmd, err = d.GetSSHCommand("echo 'export DOCKER_OPTS=\"--auth=identity --host=tcp://0.0.0.0:2376 --host=unix:///var/run/docker.sock\"' >> /etc/default/docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("Adding key to authorized-keys.d...")

	if err := drivers.AddPublicKeyToAuthorizedHosts(d, "/.docker/authorized-keys.d"); err != nil {
		return err
	}

	cmd, err = d.GetSSHCommand("start docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) createSSHKey() (*godo.Key, error) {
	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return nil, err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return nil, err
	}

	createRequest := &godo.KeyCreateRequest{
		Name:      d.MachineName,
		PublicKey: string(publicKey),
	}

	key, _, err := d.getClient().Keys.Create(createRequest)
	if err != nil {
		return key, err
	}

	return key, nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	if d.IPAddress == "" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	droplet, _, err := d.getClient().Droplets.Get(d.DropletID)
	if err != nil {
		return state.Error, err
	}
	switch droplet.Droplet.Status {
	case "new":
		return state.Starting, nil
	case "active":
		return state.Running, nil
	case "off":
		return state.Stopped, nil
	}
	return state.None, nil
}

func (d *Driver) Start() error {
	_, _, err := d.getClient().DropletActions.PowerOn(d.DropletID)
	return err
}

func (d *Driver) Stop() error {
	_, _, err := d.getClient().DropletActions.Shutdown(d.DropletID)
	return err
}

func (d *Driver) Remove() error {
	client := d.getClient()
	if resp, err := client.Keys.DeleteByID(d.SSHKeyID); err != nil {
		if resp.StatusCode == 404 {
			log.Infof("Digital Ocean SSH key doesn't exist, assuming it is already deleted")
		} else {
			return err
		}
	}
	if resp, err := client.Droplets.Delete(d.DropletID); err != nil {
		if resp.StatusCode == 404 {
			log.Infof("Digital Ocean droplet doesn't exist, assuming it is already deleted")
		} else {
			return err
		}
	}
	return nil
}

func (d *Driver) Restart() error {
	_, _, err := d.getClient().DropletActions.Reboot(d.DropletID)
	return err
}

func (d *Driver) Kill() error {
	_, _, err := d.getClient().DropletActions.PowerOff(d.DropletID)
	return err
}

func (d *Driver) Upgrade() error {
	sshCmd, err := d.GetSSHCommand("apt-get update && apt-get install lxc-docker")
	if err != nil {
		return err
	}
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	if err := sshCmd.Run(); err != nil {
		return fmt.Errorf("%s", err)
	}
	return nil
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return ssh.GetSSHCommand(d.IPAddress, 22, "root", d.sshKeyPath(), args...), nil
}

func (d *Driver) getClient() *godo.Client {
	t := &oauth.Transport{
		Token: &oauth.Token{AccessToken: d.AccessToken},
	}

	return godo.NewClient(t.Client())
}

func (d *Driver) sshKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

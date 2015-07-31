// Copyright (C) 2015  Nicolas Lamirault <nicolas.lamirault@gmail.com>

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package scaleway

import (
	//"errors"
	"fmt"
	//"os"
	// "os/exec"
	"path/filepath"
	"time"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"

	"github.com/codegangsta/cli"
	"github.com/nlamirault/go-scaleway/api"
)

const (
	driverName      = "scaleway"
	dockerConfigDir = "/etc/docker"
)

type Driver struct {
	Id             string
	MachineName    string
	SSHUser        string
	SSHPort        int
	UserId         string
	Token          string
	Organization   string
	Image          string
	Volumes        string
	IPAddress      string
	CaCertPath     string
	PrivateKeyPath string
	storePath      string
}

func init() {
	drivers.Register(driverName, &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   "scaleway-userid",
			Usage:  "Scaleway UserID",
			EnvVar: "SCALEWAY_USERID",
		},
		cli.StringFlag{
			Name:   "scaleway-token",
			Usage:  "Scaleway Token",
			EnvVar: "SCALEWAY_TOKEN",
		},
		cli.StringFlag{
			Name:   "scaleway-organization",
			Usage:  "Organization identifier",
			EnvVar: "SCALEWAY_ORGANIZATION",
		},
		cli.StringFlag{
			Name:   "scaleway-image",
			Usage:  "Image identifier",
			EnvVar: "SCALEWAY_IMAGE",
		},
		cli.StringFlag{
			Name:   "scaleway-volumes",
			Usage:  "Volumes identifier",
			EnvVar: "SCALEWAY_VOLUMES",
		},
	}
}

// NewDriver creates a Driver with the specified storePath.
func NewDriver(machineName string, storePath string, caCert string,
	privateKey string) (drivers.Driver, error) {
	return &Driver{
		MachineName:    machineName,
		storePath:      storePath,
		CaCertPath:     caCert,
		PrivateKeyPath: privateKey,
	}, nil
}

// SetConfigFromFlags initializes the driver based on the command line flags.
func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.UserId = flags.String("scaleway-userid")
	d.Token = flags.String("scaleway-token")
	d.Organization = flags.String("scaleway-organization")
	d.Image = flags.String("scaleway-image")
	d.Volumes = flags.String("scaleway-volumes")
	if d.UserId == "" {
		return fmt.Errorf("scaleway driver requires the --scaleway-userid option")
	}
	if d.Token == "" {
		return fmt.Errorf("scaleway driver requires the --scaleway-token option")
	}
	if d.Organization == "" {
		return fmt.Errorf("scaleway driver requires the --scaleway-organization option")
	}
	if d.Image == "" {
		return fmt.Errorf("scaleway driver requires the --scaleway-image option")
	}
	return nil
}

// DriverName returns the name of the driver.
func (d *Driver) DriverName() string {
	return driverName
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
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = 22
	}

	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "root"
	}

	return d.SSHUser
}

// func (d *Driver) GetDockerConfigDir() string {
// 	return dockerConfigDir
// }

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Create() error {
	log.Infof("[Scaleway] Creating instance...")
	client := d.getClient()
	response, err := client.CreateServer(
		d.MachineName, d.Image)
	if err != nil {
		return err
	}
	d.Id = response.Server.ID
	log.Debugf("[Scaleway] ServerID %s", d.Id)

	log.Debugf("[Scaleway] Create SSH key")
	err = d.createSSHKey()
	if err != nil {
		return err
	}
	log.Debugf("[Scaleway] Upload SSH key")
	if _, err = client.UploadPublicKey(d.publicSSHKeyPath()); err != nil {
		return err
	}

	if err = d.Start(); err != nil {
		return err
	}
	log.Debugf("[Scaleway] Waiting server ready .......")
	if err := d.waitForServerState(state.Running); err != nil {
		return err
	}
	log.Debugf("[Scaleway] Waiting SSH .......")
	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", d.IPAddress, 22)); err != nil {
		return err
	}

	time.Sleep(10 * time.Second)
	d.setupHostname()
	//d.installDocker()
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
	if d.IPAddress == "" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	log.Debugf("[Scaleway] Retrieving state server %s", d.Id)
	client := d.getClient()
	response, err := client.GetServer(d.Id)
	if err != nil {
		return state.Error, err
	}
	return getServerState(response.Server.State), nil
}

func (d *Driver) Start() error {
	log.Infof("[Scaleway] Starting instance...")
	client := d.getClient()
	if _, err := client.PerformServerAction(d.Id, "poweron"); err != nil {
		return err
	}
	d.waitForServerState(state.Running)
	return nil
}

func (d *Driver) Stop() error {
	log.Infof("[Scaleway] Stopping instance...")
	client := d.getClient()
	if _, err := client.PerformServerAction(d.Id, "poweroff"); err != nil {
		return err
	}
	d.waitForServerState(state.Stopped)
	return nil
}

func (d *Driver) Remove() error {
	log.Infof("[Scaleway] Removing instance... ")
	client := d.getClient()
	if err := client.DeleteServer(d.Id); err != nil {
		return err
	}
	d.waitForServerState(state.Stopped)
	return nil
}

func (d *Driver) Restart() error {
	log.Infof("[Scaleway] Rebooting instance...")
	client := d.getClient()
	if _, err := client.PerformServerAction(d.Id, "reboot"); err != nil {
		return err
	}
	d.waitForServerState(state.Running)
	return nil
}

func (d *Driver) Kill() error {
	return d.Stop()
}

// func (d *Driver) Upgrade() error {
// 	log.Debugf("[Scaleway] Upgrading Docker")
// 	_, err := drivers.RunSSHCommandFromDriver(d,
// 		"sudo apt-get update && apt-get install --upgrade lxc-docker")
// 	return err
// }

func (d *Driver) setupHostname() error {
	log.Debugf("[Scaleway] Setting hostname: %s", d.MachineName)
	_, err := drivers.RunSSHCommandFromDriver(d,
		fmt.Sprintf(
			"echo \"127.0.0.1 %s\" | sudo tee -a /etc/hosts && sudo hostname %s && echo \"%s\" | sudo tee /etc/hostname",
			d.MachineName,
			d.MachineName,
			d.MachineName,
		))
	return err
}

// func (d *Driver) installDocker() error {
// 	_, err := drivers.RunSSHCommandFromDriver(d,
// 		"if [ ! -e /usr/bin/docker ]; then curl -sL https://get.docker.com | sh -; fi")
// 	return err
// }

// func (d *Driver) StartDocker() error {
// 	log.Debug("Starting Docker...")
// 	_, err := drivers.RunSSHCommandFromDriver(d,
// 		"sudo service docker start")
// 	return err
// }

// func (d *Driver) StopDocker() error {
// 	log.Debug("Stopping Docker...")
// 	_, err := drivers.RunSSHCommandFromDriver(d,
// 		"sudo service docker stop")
// 	return err
// }

func (d *Driver) setMachineNameIfNotSet() {
}

func (d *Driver) createSSHKey() error {
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}
	return nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) getClient() *api.ScalewayClient {
	return api.NewClient(d.Token, d.UserId, d.Organization)

}

func getServerState(status string) state.State {
	switch status {
	case "stopped":
		return state.Stopped
	case "stopping":
		return state.Stopping
	case "starting":
		return state.Starting
	case "running":
		return state.Running
	}
	return state.None
}

func (d *Driver) waitForServerState(serverState state.State) error {
	client := d.getClient()
	for {
		response, err := client.GetServer(d.Id)
		if err != nil {
			return err
		}
		status := getServerState(response.Server.State)
		log.Infof("[Scaleway] Waiting server state %s. Currently : %s",
			serverState, status)
		if status == serverState {
			if status == state.Running {
				log.Infof("[Scaleway] Server %s is running. Set IP address",
					d.Id)
				d.IPAddress = response.Server.PublicIP.Address
			}
			break
		}
		time.Sleep(5 * time.Second)
	}
	return nil

}

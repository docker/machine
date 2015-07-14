package centurylinkcloud

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/CenturyLinkLabs/clcgo"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
)

const serverRequirementsWarning = "CenturyLink Cloud requires that machine names are 2 to 6 characters long"

func getCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "CENTURYLINKCLOUD_USERNAME",
			Name:   "centurylinkcloud-username",
			Usage:  "CenturyLink Cloud Username",
		},
		cli.StringFlag{
			EnvVar: "CENTURYLINKCLOUD_PASSWORD",
			Name:   "centurylinkcloud-password",
			Usage:  "CenturyLink Cloud Password",
		},
		cli.StringFlag{
			EnvVar: "CENTURYLINKCLOUD_SERVER_NAME",
			Name:   "centurylinkcloud-server-name",
			Usage:  "CenturyLink Cloud Server Name",
		},
		cli.StringFlag{
			EnvVar: "CENTURYLINKCLOUD_GROUP_ID",
			Name:   "centurylinkcloud-group-id",
			Usage:  "CenturyLink Cloud Group ID",
		},
		cli.StringFlag{
			EnvVar: "CENTURYLINKCLOUD_SOURCE_SERVER_ID",
			Name:   "centurylinkcloud-source-server-id",
			Usage:  "CenturyLink Cloud Source Server ID",
			Value:  "UBUNTU-14-64-TEMPLATE",
		},
		cli.IntFlag{
			EnvVar: "CENTURYLINKCLOUD_CPU",
			Name:   "centurylinkcloud-cpu",
			Usage:  "CenturyLink Cloud CPU Count",
			Value:  1,
		},
		cli.IntFlag{
			EnvVar: "CENTURYLINKCLOUD_MEMORYGB",
			Name:   "centurylinkcloud-memory-gb",
			Usage:  "CenturyLink Cloud Memory GB",
			Value:  2,
		},
	}
}

func (d *Driver) PreCreateCheck() error {
	i := len(d.MachineName)
	if i < 2 {
		return errors.New(serverRequirementsWarning)
	}

	return nil
}

func (d *Driver) Create() error {
	c, err := d.getClientWithPersistence()
	if err != nil {
		return err
	}

	s, err := d.createServer(c)
	if err != nil {
		return err
	}

	if err := d.addPublicIPAddress(c, &s); err != nil {
		return err
	}

	if err = d.generateAndWriteSSHKey(c, s); err != nil {
		return err
	}

	return nil
}

func (d *Driver) createServer(c *clcgo.Client) (clcgo.Server, error) {
	log.Infof("Creating server...")

	name := d.MachineName[:6]
	if name != d.MachineName {
		log.Warnf(
			"%s. Your server name has been truncated to: %s",
			serverRequirementsWarning, name,
		)
	}

	s := clcgo.Server{
		Name:           name,
		GroupID:        d.GroupID,
		SourceServerID: d.SourceServerID,
		CPU:            d.CPU,
		MemoryGB:       d.MemoryGB,
		Type:           "standard",
	}

	st, err := c.SaveEntity(&s)
	if err != nil {
		return s, logAndReturnError(err)
	}

	for !st.HasSucceeded() {
		time.Sleep(time.Second * statusWaitSeconds)
		log.Debugf("Checking status...")
		if err := c.GetEntity(&st); err != nil {
			return s, err
		}
	}

	if err = c.GetEntity(&s); err != nil {
		return s, err
	}
	d.ServerID = s.ID
	log.Infof("Server '%s' is provisioned", s.Name)

	return s, nil
}

func (d Driver) addPublicIPAddress(c *clcgo.Client, s *clcgo.Server) error {
	log.Infof("Adding public IP address...")

	ports := []clcgo.Port{
		{Protocol: "TCP", Port: 22},   // SSH
		{Protocol: "TCP", Port: 2376}, // Docker
	}
	a := clcgo.PublicIPAddress{Server: *s, Ports: ports}
	st, err := c.SaveEntity(&a)
	if err != nil {
		return logAndReturnError(err)
	}
	for !st.HasSucceeded() {
		time.Sleep(time.Second * statusWaitSeconds)
		log.Debugf("Checking status...")
		err = c.GetEntity(&st)
		if err != nil {
			return err
		}
	}

	if err := c.GetEntity(s); err != nil {
		return err
	}
	ip := publicIPFromServer(*s)
	if ip == "" {
		return errors.New("could not find an IP Address for the server")
	}

	log.Infof("IP Address '%s' has been provisioned", ip)

	return nil
}

func (d Driver) generateAndWriteSSHKey(c *clcgo.Client, s clcgo.Server) error {
	cr := clcgo.Credentials{Server: s}
	if err := c.GetEntity(&cr); err != nil {
		return err
	}

	hostname, err := d.GetSSHHostname()
	if err != nil {
		return err
	}
	port, err := d.GetSSHPort()
	if err != nil {
		return err
	}

	auth := ssh.Auth{Passwords: []string{cr.Password}}
	sshClient, err := ssh.NewNativeClient(d.GetSSHUsername(), hostname, port, &auth)
	if err != nil {
		return err
	}

	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	keyPath := d.GetSSHKeyPath() + ".pub"
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return err
	}

	log.Info("Adding public key to authorized_keys")
	cmd := fmt.Sprintf(`echo "%s" >> ~/.ssh/authorized_keys`, string(key))
	if _, err = sshClient.Output(cmd); err != nil {
		return err
	}

	return nil
}

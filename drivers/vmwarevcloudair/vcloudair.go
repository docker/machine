/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwarevcloudair

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"

	"github.com/vmware/govcloudair"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/docker/utils"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

type Driver struct {
	UserName     string
	UserPassword string
	ComputeID    string
	VDCID        string
	OrgVDCNet    string
	EdgeGateway  string
	PublicIP     string
	Catalog      string
	CatalogItem  string
	MachineName  string
	SSHPort      int
	DockerPort   int
	Provision    bool
	CPUCount     int
	MemorySize   int

	VAppID    string
	storePath string
}

type CreateFlags struct {
	UserName     *string
	UserPassword *string
	ComputeID    *string
	VDCID        *string
	OrgVDCNet    *string
	EdgeGateway  *string
	PublicIP     *string
	Catalog      *string
	CatalogItem  *string
	Name         *string
	SSHPort      *int
	DockerPort   *int
	Provision    *bool
	CPUCount     *int
	MemorySize   *int
}

func init() {
	drivers.Register("vmwarevcloudair", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_USERNAME",
			Name:   "vmwarevcloudair-username",
			Usage:  "vCloud Air username",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_PASSWORD",
			Name:   "vmwarevcloudair-password",
			Usage:  "vCloud Air password",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_COMPUTEID",
			Name:   "vmwarevcloudair-computeid",
			Usage:  "vCloud Air Compute ID (if using Dedicated Cloud)",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_VDCID",
			Name:   "vmwarevcloudair-vdcid",
			Usage:  "vCloud Air VDC ID",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_ORGVDCNETWORK",
			Name:   "vmwarevcloudair-orgvdcnetwork",
			Usage:  "vCloud Air Org VDC Network (Default is <vdcid>-default-routed)",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_EDGEGATEWAY",
			Name:   "vmwarevcloudair-edgegateway",
			Usage:  "vCloud Air Org Edge Gateway (Default is <vdcid>)",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_PUBLICIP",
			Name:   "vmwarevcloudair-publicip",
			Usage:  "vCloud Air Org Public IP to use",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_CATALOG",
			Name:   "vmwarevcloudair-catalog",
			Usage:  "vCloud Air Catalog (default is Public Catalog)",
			Value:  "Public Catalog",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_CATALOGITEM",
			Name:   "vmwarevcloudair-catalogitem",
			Usage:  "vCloud Air Catalog Item (default is Ubuntu Precise)",
			Value:  "Ubuntu Server 12.04 LTS (amd64 20140927)",
		},

		// BoolTFlag is true by default.
		cli.BoolTFlag{
			EnvVar: "VCLOUDAIR_PROVISION",
			Name:   "vmwarevcloudair-provision",
			Usage:  "vCloud Air Install Docker binaries (default is true)",
		},

		cli.IntFlag{
			EnvVar: "VCLOUDAIR_CPU_COUNT",
			Name:   "vmwarevcloudair-cpu-count",
			Usage:  "vCloud Air VM Cpu Count (default 1)",
			Value:  1,
		},
		cli.IntFlag{
			EnvVar: "VCLOUDAIR_MEMORY_SIZE",
			Name:   "vmwarevcloudair-memory-size",
			Usage:  "vCloud Air VM Memory Size in MB (default 2048)",
			Value:  2048,
		},
		cli.IntFlag{
			EnvVar: "VCLOUDAIR_SSH_PORT",
			Name:   "vmwarevcloudair-ssh-port",
			Usage:  "vCloud Air SSH port",
			Value:  22,
		},
		cli.IntFlag{
			EnvVar: "VCLOUDAIR_DOCKER_PORT",
			Name:   "vmwarevcloudair-docker-port",
			Usage:  "vCloud Air Docker port",
			Value:  2376,
		},
	}
}

func NewDriver(machineName string, storePath string) (drivers.Driver, error) {
	driver := &Driver{MachineName: machineName, storePath: storePath}
	return driver, nil
}

// Driver interface implementation
func (driver *Driver) DriverName() string {
	return "vmwarevcloudair"
}

func (driver *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {

	driver.UserName = flags.String("vmwarevcloudair-username")
	driver.UserPassword = flags.String("vmwarevcloudair-password")
	driver.VDCID = flags.String("vmwarevcloudair-vdcid")
	driver.PublicIP = flags.String("vmwarevcloudair-publicip")

	// Check for required Params
	if driver.UserName == "" || driver.UserPassword == "" || driver.VDCID == "" || driver.PublicIP == "" {
		return fmt.Errorf("Please specify vcloudair mandatory params using options: -vmwarevcloudair-username -vmwarevcloudair-password -vmwarevcloudair-vdcid and -vmwarevcloudair-publicip")
	}

	// If ComputeID is not set we're using a VPC, hence setting ComputeID = VDCID
	if flags.String("vmwarevcloudair-computeid") == "" {
		driver.ComputeID = flags.String("vmwarevcloudair-vdcid")
	} else {
		driver.ComputeID = flags.String("vmwarevcloudair-computeid")
	}

	// If the Org VDC Network is empty, set it to the default routed network.
	if flags.String("vmwarevcloudair-orgvdcnetwork") == "" {
		driver.OrgVDCNet = flags.String("vmwarevcloudair-vdcid") + "-default-routed"
	} else {
		driver.OrgVDCNet = flags.String("vmwarevcloudair-orgvdcnetwork")
	}

	// If the Edge Gateway is empty, just set it to the default edge gateway.
	if flags.String("vmwarevcloudair-edgegateway") == "" {
		driver.EdgeGateway = flags.String("vmwarevcloudair-vdcid")
	} else {
		driver.EdgeGateway = flags.String("vmwarevcloudair-edgegateway")
	}

	driver.Catalog = flags.String("vmwarevcloudair-catalog")
	driver.CatalogItem = flags.String("vmwarevcloudair-catalogitem")

	driver.DockerPort = flags.Int("vmwarevcloudair-docker-port")
	driver.SSHPort = flags.Int("vmwarevcloudair-ssh-port")
	driver.Provision = flags.Bool("vmwarevcloudair-provision")
	driver.CPUCount = flags.Int("vmwarevcloudair-cpu-count")
	driver.MemorySize = flags.Int("vmwarevcloudair-memory-size")

	return nil
}

func (d *Driver) GetURL() (string, error) {
	return fmt.Sprintf("tcp://%s:%d", d.PublicIP, d.DockerPort), nil
}

func (d *Driver) GetIP() (string, error) {
	return d.PublicIP, nil
}

func (d *Driver) GetState() (state.State, error) {

	p, err := govcloudair.NewClient()
	if err != nil {
		return state.Error, err
	}

	log.Infof("Connecting to vCloud Air to fetch vApp Status...")
	// Authenticate to vCloud Air
	v, err := p.Authenticate(d.UserName, d.UserPassword, d.ComputeID, d.VDCID)
	if err != nil {
		return state.Error, err
	}

	vapp, err := v.FindVAppByID(d.VAppID)
	if err != nil {
		return state.Error, err
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return state.Error, err
	}

	if err = p.Disconnect(); err != nil {
		return state.Error, err
	}

	switch status {
	case "POWERED_ON":
		return state.Running, nil
	case "POWERED_OFF":
		return state.Stopped, nil
	}
	return state.None, nil

}

func (d *Driver) Create() error {

	key, err := d.createSSHKey()
	if err != nil {
		return err
	}

	p, err := govcloudair.NewClient()
	if err != nil {
		return err
	}

	log.Infof("Connecting to vCloud Air...")
	// Authenticate to vCloud Air
	v, err := p.Authenticate(d.UserName, d.UserPassword, d.ComputeID, d.VDCID)
	if err != nil {
		return err
	}

	// Find VDC Network
	net, err := v.FindVDCNetwork(d.OrgVDCNet)
	if err != nil {
		return err
	}

	// Find our Edge Gateway
	edge, err := v.FindEdgeGateway(d.EdgeGateway)
	if err != nil {
		return err
	}

	// Get the Org our VDC belongs to
	org, err := v.GetVDCOrg()
	if err != nil {
		return err
	}

	// Find our Catalog
	cat, err := org.FindCatalog(d.Catalog)
	if err != nil {
		return err
	}

	// Find our Catalog Item
	cati, err := cat.FindCatalogItem(d.CatalogItem)
	if err != nil {
		return err
	}

	// Fetch the vApp Template in the Catalog Item
	vapptemplate, err := cati.GetVAppTemplate()
	if err != nil {
		return err
	}

	// Create a new empty vApp
	vapp := govcloudair.NewVApp(p)

	log.Infof("Creating a new vApp: %s...", d.MachineName)
	// Compose the vApp with ComposeVApp
	task, err := vapp.ComposeVApp(net, vapptemplate, d.MachineName, "Container Host created with Docker Host")
	if err != nil {
		return err
	}

	// Wait for the creation to be completed
	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	task, err = vapp.ChangeCPUcount(d.CPUCount)
	if err != nil {
		return err
	}

	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	task, err = vapp.ChangeMemorySize(d.MemorySize)
	if err != nil {
		return err
	}

	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	sshCustomScript := "echo \"" + strings.TrimSpace(key) + "\" > /root/.ssh/authorized_keys"

	task, err = vapp.RunCustomizationScript(d.MachineName, sshCustomScript)
	if err != nil {
		return err
	}

	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	task, err = vapp.PowerOn()
	if err != nil {
		return err
	}

	log.Infof("Waiting for the VM to power on and run the customization script...")

	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	log.Infof("Creating NAT and Firewall Rules on %s...", d.EdgeGateway)
	task, err = edge.Create1to1Mapping(vapp.VApp.Children.VM[0].NetworkConnectionSection.NetworkConnection.IPAddress, d.PublicIP, d.MachineName)
	if err != nil {
		return err
	}

	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	log.Infof("Waiting for SSH...")

	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", d.PublicIP, d.SSHPort)); err != nil {
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

	connTest := "ping -c 3 www.google.com >/dev/null 2>&1 && ( echo \"Connectivity and DNS tests passed.\" ) || ( echo \"Connectivity and DNS tests failed, trying to add Nameserver to resolv.conf\"; echo \"nameserver 8.8.8.8\" >> /etc/resolv.conf )"

	log.Debugf("Connectivity and DNS sanity test...")
	cmd, err = d.GetSSHCommand(connTest)
	if err != nil {
		return err
	}

	if err := cmd.Run(); err != nil {
		return err
	}

	if d.Provision {
		dockerInstall := "curl -sSL https://get.docker.com/ | sudo sh"

		log.Infof("Installing Docker...")

		cmd, err = d.GetSSHCommand(dockerInstall)
		if err != nil {
			return err
		}

		if err = cmd.Run(); err != nil {
			return err
		}

		log.Debugf("Stopping Docker")

		cmd, err = d.GetSSHCommand("stop docker")
		if err != nil {
			return err
		}
		if err := cmd.Run(); err != nil {
			return err
		}

		log.Debugf("Replacing docker binary with a version that supports identity authentication")
		cmd, err = d.GetSSHCommand("curl -sS https://bfirsh.s3.amazonaws.com/docker/docker-1.3.1-dev-identity-auth > /usr/bin/docker")
		if err != nil {
			return err
		}
		if err := cmd.Run(); err != nil {
			return err
		}

		dockerListen := "echo 'export DOCKER_OPTS=\"--auth=identity --host=tcp://0.0.0.0:" + strconv.Itoa(d.DockerPort) + "\"' >> /etc/default/docker"

		log.Infof("Updating /etc/default/docker to listen on all interfaces...")

		cmd, err = d.GetSSHCommand(dockerListen)
		if err != nil {
			return err
		}

		if err = cmd.Run(); err != nil {
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

	}

	log.Debugf("Disconnecting from vCloud Air...")

	if err = p.Disconnect(); err != nil {
		return err
	}

	// Set VAppID with ID of the created VApp
	d.VAppID = vapp.VApp.ID

	return nil

}

func (d *Driver) Remove() error {

	p, err := govcloudair.NewClient()
	if err != nil {
		return err
	}

	log.Infof("Connecting to vCloud Air...")
	// Authenticate to vCloud Air
	v, err := p.Authenticate(d.UserName, d.UserPassword, d.ComputeID, d.VDCID)
	if err != nil {
		return err
	}

	// Find our Edge Gateway
	edge, err := v.FindEdgeGateway(d.EdgeGateway)
	if err != nil {
		return err
	}

	vapp, err := v.FindVAppByID(d.VAppID)
	if err != nil {
		log.Infof("Can't find the vApp, assuming it was deleted already...")
		return nil
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return err
	}

	log.Infof("Removing NAT and Firewall Rules on %s...", d.EdgeGateway)
	task, err := edge.Remove1to1Mapping(vapp.VApp.Children.VM[0].NetworkConnectionSection.NetworkConnection.IPAddress, d.PublicIP)
	if err != nil {
		return err
	}
	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	if status == "POWERED_ON" {
		// If it's powered on, power it off before deleting
		log.Infof("Powering Off %s...", d.MachineName)
		task, err = vapp.PowerOff()
		if err != nil {
			return err
		}
		if err = task.WaitTaskCompletion(); err != nil {
			return err
		}

	}

	log.Debugf("Undeploying %s...", d.MachineName)
	task, err = vapp.Undeploy()
	if err != nil {
		return err
	}
	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	log.Infof("Deleting %s...", d.MachineName)
	task, err = vapp.Delete()
	if err != nil {
		return err
	}
	if err = task.WaitTaskCompletion(); err != nil {
		return err
	}

	if err = p.Disconnect(); err != nil {
		return err
	}

	return nil

}

func (d *Driver) Start() error {

	p, err := govcloudair.NewClient()
	if err != nil {
		return err
	}

	log.Infof("Connecting to vCloud Air...")
	// Authenticate to vCloud Air
	v, err := p.Authenticate(d.UserName, d.UserPassword, d.ComputeID, d.VDCID)
	if err != nil {
		return err
	}

	vapp, err := v.FindVAppByID(d.VAppID)
	if err != nil {
		return err
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return err
	}

	if status == "POWERED_OFF" {
		log.Infof("Starting %s...", d.MachineName)
		task, err := vapp.PowerOn()
		if err != nil {
			return err
		}
		if err = task.WaitTaskCompletion(); err != nil {
			return err
		}

	}

	if err = p.Disconnect(); err != nil {
		return err
	}

	return nil

}

func (d *Driver) Stop() error {

	p, err := govcloudair.NewClient()
	if err != nil {
		return err
	}

	log.Infof("Connecting to vCloud Air...")
	// Authenticate to vCloud Air
	v, err := p.Authenticate(d.UserName, d.UserPassword, d.ComputeID, d.VDCID)
	if err != nil {
		return err
	}

	vapp, err := v.FindVAppByID(d.VAppID)
	if err != nil {
		return err
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return err
	}

	if status == "POWERED_ON" {
		log.Infof("Shutting down %s...", d.MachineName)
		task, err := vapp.Shutdown()
		if err != nil {
			return err
		}
		if err = task.WaitTaskCompletion(); err != nil {
			return err
		}

	}

	if err = p.Disconnect(); err != nil {
		return err
	}

	return nil

}

func (d *Driver) Restart() error {

	p, err := govcloudair.NewClient()
	if err != nil {
		return err
	}

	log.Infof("Connecting to vCloud Air...")
	// Authenticate to vCloud Air
	v, err := p.Authenticate(d.UserName, d.UserPassword, d.ComputeID, d.VDCID)
	if err != nil {
		return err
	}

	vapp, err := v.FindVAppByID(d.VAppID)
	if err != nil {
		return err
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return err
	}

	if status == "POWERED_ON" {
		// If it's powered on, restart the machine
		log.Infof("Restarting %s...", d.MachineName)
		task, err := vapp.Reset()
		if err != nil {
			return err
		}
		if err = task.WaitTaskCompletion(); err != nil {
			return err
		}

	} else {
		// If it's not powered on, start it.
		log.Infof("Docker host %s is powered off, powering it back on...", d.MachineName)
		task, err := vapp.PowerOn()
		if err != nil {
			return err
		}
		if err = task.WaitTaskCompletion(); err != nil {
			return err
		}

	}

	if err = p.Disconnect(); err != nil {
		return err
	}

	return nil

}

func (d *Driver) Kill() error {

	p, err := govcloudair.NewClient()
	if err != nil {
		return err
	}

	log.Infof("Connecting to vCloud Air...")
	// Authenticate to vCloud Air
	v, err := p.Authenticate(d.UserName, d.UserPassword, d.ComputeID, d.VDCID)
	if err != nil {
		return err
	}

	vapp, err := v.FindVAppByID(d.VAppID)
	if err != nil {
		return err
	}

	status, err := vapp.GetStatus()
	if err != nil {
		return err
	}

	if status == "POWERED_ON" {
		log.Infof("Stopping %s...", d.MachineName)
		task, err := vapp.PowerOff()
		if err != nil {
			return err
		}
		if err = task.WaitTaskCompletion(); err != nil {
			return err
		}

	}

	if err = p.Disconnect(); err != nil {
		return err
	}

	return nil

}

func (d *Driver) Upgrade() error {
	// Stolen from DigitalOcean ;-)
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
	return ssh.GetSSHCommand(d.PublicIP, d.SSHPort, "root", d.sshKeyPath(), args...), nil
}

// Helpers

func generateVMName() string {
	randomID := utils.TruncateID(utils.GenerateRandomID())
	return fmt.Sprintf("docker-host-%s", randomID)
}

func (d *Driver) createSSHKey() (string, error) {

	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return "", err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return "", err
	}

	return string(publicKey), nil
}

func (d *Driver) sshKeyPath() string {
	return path.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

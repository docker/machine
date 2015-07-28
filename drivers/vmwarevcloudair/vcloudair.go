/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwarevcloudair

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/vmware/govcloudair"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
)

type Driver struct {
	*drivers.BaseDriver
	UserName     string
	UserPassword string
	ComputeID    string
	VDCID        string
	OrgVDCNet    string
	EdgeGateway  string
	PublicIP     string
	Catalog      string
	CatalogItem  string
	DockerPort   int
	Provision    bool
	CPUCount     int
	MemorySize   int
	VAppID       string
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
			Name:   "username",
			Usage:  "vCloud Air username",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_PASSWORD",
			Name:   "password",
			Usage:  "vCloud Air password",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_COMPUTEID",
			Name:   "computeid",
			Usage:  "vCloud Air Compute ID (if using Dedicated Cloud)",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_VDCID",
			Name:   "vdcid",
			Usage:  "vCloud Air VDC ID",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_ORGVDCNETWORK",
			Name:   "orgvdcnetwork",
			Usage:  "vCloud Air Org VDC Network (Default is <vdcid>-default-routed)",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_EDGEGATEWAY",
			Name:   "edgegateway",
			Usage:  "vCloud Air Org Edge Gateway (Default is <vdcid>)",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_PUBLICIP",
			Name:   "publicip",
			Usage:  "vCloud Air Org Public IP to use",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_CATALOG",
			Name:   "catalog",
			Usage:  "vCloud Air Catalog (default is Public Catalog)",
			Value:  "Public Catalog",
		},
		cli.StringFlag{
			EnvVar: "VCLOUDAIR_CATALOGITEM",
			Name:   "catalogitem",
			Usage:  "vCloud Air Catalog Item (default is Ubuntu Precise)",
			Value:  "Ubuntu Server 12.04 LTS (amd64 20150127)",
		},

		// BoolTFlag is true by default.
		cli.BoolTFlag{
			EnvVar: "VCLOUDAIR_PROVISION",
			Name:   "provision",
			Usage:  "vCloud Air Install Docker binaries (default is true)",
		},

		cli.IntFlag{
			EnvVar: "VCLOUDAIR_CPU_COUNT",
			Name:   "cpu-count",
			Usage:  "vCloud Air VM Cpu Count (default 1)",
			Value:  1,
		},
		cli.IntFlag{
			EnvVar: "VCLOUDAIR_MEMORY_SIZE",
			Name:   "memory-size",
			Usage:  "vCloud Air VM Memory Size in MB (default 2048)",
			Value:  2048,
		},
		cli.IntFlag{
			EnvVar: "VCLOUDAIR_SSH_PORT",
			Name:   "ssh-port",
			Usage:  "vCloud Air SSH port",
			Value:  22,
		},
		cli.IntFlag{
			EnvVar: "VCLOUDAIR_DOCKER_PORT",
			Name:   "docker-port",
			Usage:  "vCloud Air Docker port",
			Value:  2376,
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	inner := drivers.NewBaseDriver(machineName, storePath, caCert, privateKey)
	return &Driver{BaseDriver: inner}, nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// Driver interface implementation
func (d *Driver) DriverName() string {
	return ""
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {

	d.UserName = flags.String("username")
	d.UserPassword = flags.String("password")
	d.VDCID = flags.String("vdcid")
	d.PublicIP = flags.String("publicip")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")

	// Check for required Params
	if d.UserName == "" || d.UserPassword == "" || d.VDCID == "" || d.PublicIP == "" {
		return fmt.Errorf("Please specify vcloudair mandatory params using options: -vmwarevcloudair-username -vmwarevcloudair-password -vmwarevcloudair-vdcid and -vmwarevcloudair-publicip")
	}

	// If ComputeID is not set we're using a VPC, hence setting ComputeID = VDCID
	if flags.String("computeid") == "" {
		d.ComputeID = flags.String("vdcid")
	} else {
		d.ComputeID = flags.String("computeid")
	}

	// If the Org VDC Network is empty, set it to the default routed network.
	if flags.String("orgvdcnetwork") == "" {
		d.OrgVDCNet = flags.String("vdcid") + "-default-routed"
	} else {
		d.OrgVDCNet = flags.String("orgvdcnetwork")
	}

	// If the Edge Gateway is empty, just set it to the default edge gateway.
	if flags.String("edgegateway") == "" {
		d.EdgeGateway = flags.String("-vdcid")
	} else {
		d.EdgeGateway = flags.String("edgegateway")
	}

	d.Catalog = flags.String("catalog")
	d.CatalogItem = flags.String("catalogitem")

	d.DockerPort = flags.Int("docker-port")
	d.SSHUser = "root"
	d.SSHPort = flags.Int("ssh-port")
	d.Provision = flags.Bool("provision")
	d.CPUCount = flags.Int("cpu-count")
	d.MemorySize = flags.Int("memory-size")

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

	log.Debug("Connecting to vCloud Air to fetch vApp Status...")
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

func (d *Driver) PreCreateCheck() error {
	return nil
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

	log.Debugf("Disconnecting from vCloud Air...")

	if err = p.Disconnect(); err != nil {
		return err
	}

	// Set VAppID with ID of the created VApp
	d.VAppID = vapp.VApp.ID

	d.IPAddress, err = d.GetIP()
	return err
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

	d.IPAddress, err = d.GetIP()
	return err
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

	d.IPAddress = ""

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

	d.IPAddress, err = d.GetIP()
	return err
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

	d.IPAddress = ""

	return nil
}

// Helpers

func generateVMName() string {
	randomID := utils.TruncateID(utils.GenerateRandomID())
	return fmt.Sprintf("docker-host-%s", randomID)
}

func (d *Driver) createSSHKey() (string, error) {
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return "", err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return "", err
	}

	return string(publicKey), nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

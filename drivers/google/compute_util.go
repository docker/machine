package google

import (
	"fmt"
	"io/ioutil"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/ssh"
	raw "google.golang.org/api/compute/v1"
)

// ComputeUtil is used to wrap the raw GCE API code and store common parameters.
type ComputeUtil struct {
	zone         string
	instanceName string
	userName     string
	project      string
	service      *raw.Service
	zoneURL      string
	globalURL    string
	ipAddress    string
}

const (
	apiURL = "https://www.googleapis.com/compute/v1/projects/"
	//imageName          = "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/container-vm-v20150129"
	imageName          = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-1404-trusty-v20150128"
	firewallRule       = "docker-machines"
	port               = "2376"
	firewallTargetTag  = "docker-machine"
	dockerStartCommand = "sudo service docker start"
	dockerStopCommand  = "sudo service docker stop"
)

const ()

// NewComputeUtil creates and initializes a ComputeUtil.
func newComputeUtil(driver *Driver) (*ComputeUtil, error) {
	service, err := newGCEService(driver.storePath)
	if err != nil {
		return nil, err
	}
	c := ComputeUtil{
		zone:         driver.Zone,
		instanceName: driver.MachineName,
		userName:     driver.UserName,
		project:      driver.Project,
		service:      service,
		zoneURL:      apiURL + driver.Project + "/zones/" + driver.Zone,
		globalURL:    apiURL + driver.Project + "/global",
	}
	return &c, nil
}

func (c *ComputeUtil) diskName() string {
	return c.instanceName + "-disk"
}

// disk returns the gce Disk.
func (c *ComputeUtil) disk() (*raw.Disk, error) {
	return c.service.Disks.Get(c.project, c.zone, c.diskName()).Do()
}

// deleteDisk deletes the persistent disk.
func (c *ComputeUtil) deleteDisk() error {
	log.Infof("Deleting disk.")
	op, err := c.service.Disks.Delete(c.project, c.zone, c.diskName()).Do()
	if err != nil {
		return err
	}
	log.Infof("Waiting for disk to delete.")
	return c.waitForRegionalOp(op.Name)
}

func (c *ComputeUtil) firewallRule() (*raw.Firewall, error) {
	return c.service.Firewalls.Get(c.project, firewallRule).Do()
}

func (c *ComputeUtil) createFirewallRule() error {
	log.Infof("Creating firewall rule.")
	rule := &raw.Firewall{
		Allowed: []*raw.FirewallAllowed{
			{
				IPProtocol: "tcp",
				Ports: []string{
					port,
				},
			},
		},
		SourceRanges: []string{
			"0.0.0.0/0",
		},
		TargetTags: []string{
			firewallTargetTag,
		},
		Name: firewallRule,
	}
	op, err := c.service.Firewalls.Insert(c.project, rule).Do()
	if err != nil {
		return err
	}
	return c.waitForGlobalOp(op.Name)
}

// instance retrieves the instance.
func (c *ComputeUtil) instance() (*raw.Instance, error) {
	return c.service.Instances.Get(c.project, c.zone, c.instanceName).Do()
}

// createInstance creates a GCE VM instance.
func (c *ComputeUtil) createInstance(d *Driver) error {
	log.Infof("Creating instance.")
	// The rule will either exist or be nil in case of an error.
	if rule, _ := c.firewallRule(); rule == nil {
		if err := c.createFirewallRule(); err != nil {
			return err
		}
	}

	instance := &raw.Instance{
		Name:        c.instanceName,
		Description: "docker host vm",
		MachineType: c.zoneURL + "/machineTypes/" + d.MachineType,
		Disks: []*raw.AttachedDisk{
			{
				Boot:       true,
				AutoDelete: false,
				Type:       "PERSISTENT",
				Mode:       "READ_WRITE",
			},
		},
		NetworkInterfaces: []*raw.NetworkInterface{
			{
				AccessConfigs: []*raw.AccessConfig{
					{Type: "ONE_TO_ONE_NAT"},
				},
				Network: c.globalURL + "/networks/default",
			},
		},
		Tags: &raw.Tags{
			Items: []string{
				firewallTargetTag,
			},
		},
	}
	disk, err := c.disk()
	if disk == nil || err != nil {
		instance.Disks[0].InitializeParams = &raw.AttachedDiskInitializeParams{
			DiskName:    c.diskName(),
			SourceImage: imageName,
			// The maximum supported disk size is 1000GB, the cast should be fine.
			DiskSizeGb: int64(d.DiskSize),
		}
	} else {
		instance.Disks[0].Source = c.zoneURL + "/disks/" + c.instanceName + "-disk"
	}
	op, err := c.service.Instances.Insert(c.project, c.zone, instance).Do()

	if err != nil {
		return err
	}
	log.Infof("Waiting for Instance...")
	if err = c.waitForRegionalOp(op.Name); err != nil {
		return err
	}

	instance, err = c.instance()
	if err != nil {
		return err
	}
	ip := instance.NetworkInterfaces[0].AccessConfigs[0].NatIP
	c.waitForSSH(ip)

	// Update the SSH Key
	sshKey, err := ioutil.ReadFile(d.publicSSHKeyPath)
	if err != nil {
		return err
	}
	log.Infof("Uploading SSH Key")
	op, err = c.service.Instances.SetMetadata(c.project, c.zone, c.instanceName, &raw.Metadata{
		Fingerprint: instance.Metadata.Fingerprint,
		Items: []*raw.MetadataItems{
			{
				Key:   "sshKeys",
				Value: c.userName + ":" + string(sshKey) + "\n",
			},
		},
	}).Do()
	if err != nil {
		return err
	}
	log.Infof("Waiting for SSH Key")
	err = c.waitForRegionalOp(op.Name)
	if err != nil {
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

	log.Debugf("Installing Docker")

	cmd, err = d.GetSSHCommand("if [ ! -e /usr/bin/docker ]; then curl -sL https://get.docker.com | sudo sh -; fi")
	if err != nil {
		return err

	}
	if err := cmd.Run(); err != nil {
		return err

	}

	return nil
}

func (c *ComputeUtil) updateDocker(d *Driver) error {
	log.Debugf("Upgrading Docker")

	cmd, err := d.GetSSHCommand("sudo apt-get update && sudo apt-get install --upgrade lxc-docker")
	if err != nil {
		return err

	}
	if err := cmd.Run(); err != nil {
		return err

	}

	return nil
}

// deleteInstance deletes the instance, leaving the persistent disk.
func (c *ComputeUtil) deleteInstance() error {
	log.Infof("Deleting instance.")
	op, err := c.service.Instances.Delete(c.project, c.zone, c.instanceName).Do()
	if err != nil {
		return err
	}
	log.Infof("Waiting for instance to delete.")
	return c.waitForRegionalOp(op.Name)
}

func (c *ComputeUtil) executeCommands(commands []string, ip, sshKeyPath string) error {
	for _, command := range commands {
		cmd := ssh.GetSSHCommand(ip, 22, c.userName, sshKeyPath, command)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error executing command: %v %v", command, err)
		}
	}
	return nil
}

func (c *ComputeUtil) waitForOp(opGetter func() (*raw.Operation, error)) error {
	for {
		op, err := opGetter()
		if err != nil {
			return err
		}
		log.Debugf("operation %q status: %s", op.Name, op.Status)
		if op.Status == "DONE" {
			if op.Error != nil {
				return fmt.Errorf("Operation error: %v", *op.Error.Errors[0])
			}
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

// waitForOp waits for the GCE Operation to finish.
func (c *ComputeUtil) waitForRegionalOp(name string) error {
	return c.waitForOp(func() (*raw.Operation, error) {
		return c.service.ZoneOperations.Get(c.project, c.zone, name).Do()
	})
}

func (c *ComputeUtil) waitForGlobalOp(name string) error {
	return c.waitForOp(func() (*raw.Operation, error) {
		return c.service.GlobalOperations.Get(c.project, name).Do()
	})
}

// waitForSSH waits for SSH to become ready on the instance.
func (c *ComputeUtil) waitForSSH(ip string) error {
	log.Infof("Waiting for SSH...")
	return ssh.WaitForTCP(fmt.Sprintf("%s:22", ip))
}

// ip retrieves and returns the external IP address of the instance.
func (c *ComputeUtil) ip() (string, error) {
	if c.ipAddress == "" {
		instance, err := c.service.Instances.Get(c.project, c.zone, c.instanceName).Do()
		if err != nil {
			return "", err
		}
		c.ipAddress = instance.NetworkInterfaces[0].AccessConfigs[0].NatIP
	}
	return c.ipAddress, nil
}

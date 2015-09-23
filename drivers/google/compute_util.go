package google

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/ssh"
	raw "google.golang.org/api/compute/v1"
)

// ComputeUtil is used to wrap the raw GCE API code and store common parameters.
type ComputeUtil struct {
	zone          string
	instanceName  string
	userName      string
	project       string
	diskTypeURL   string
	address       string
	preemptible   bool
	service       *raw.Service
	zoneURL       string
	authTokenPath string
	globalURL     string
	ipAddress     string
	SwarmMaster   bool
	SwarmHost     string
}

const (
	apiURL             = "https://www.googleapis.com/compute/v1/projects/"
	imageName          = "https://www.googleapis.com/compute/v1/projects/ubuntu-os-cloud/global/images/ubuntu-1404-trusty-v20150316"
	firewallRule       = "docker-machines"
	port               = "2376"
	firewallTargetTag  = "docker-machine"
	dockerStartCommand = "sudo service docker start"
	dockerStopCommand  = "sudo service docker stop"
)

// NewComputeUtil creates and initializes a ComputeUtil.
func newComputeUtil(driver *Driver) (*ComputeUtil, error) {
	service, err := newGCEService(driver.ResolveStorePath("."), driver.AuthTokenPath)
	if err != nil {
		return nil, err
	}
	c := ComputeUtil{
		authTokenPath: driver.AuthTokenPath,
		zone:          driver.Zone,
		instanceName:  driver.MachineName,
		userName:      driver.SSHUser,
		project:       driver.Project,
		diskTypeURL:   driver.DiskType,
		address:       driver.Address,
		preemptible:   driver.Preemptible,
		service:       service,
		zoneURL:       apiURL + driver.Project + "/zones/" + driver.Zone,
		globalURL:     apiURL + driver.Project + "/global",
		SwarmMaster:   driver.SwarmMaster,
		SwarmHost:     driver.SwarmHost,
	}
	return &c, nil
}

func (c *ComputeUtil) diskName() string {
	return c.instanceName + "-disk"
}

func (c *ComputeUtil) diskType() string {
	return apiURL + c.project + "/zones/" + c.zone + "/diskTypes/" + c.diskTypeURL
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

// staticAddress returns the external static IP address.
func (c *ComputeUtil) staticAddress() (string, error) {
	// is the address a name?
	isName, err := regexp.MatchString("[a-z]([-a-z0-9]*[a-z0-9])?", c.address)
	if err != nil {
		return "", err
	}

	if !isName {
		return c.address, nil
	}

	// resolve the address by name
	externalAddress, err := c.service.Addresses.Get(c.project, c.region(), c.address).Do()
	if err != nil {
		return "", err
	}

	return externalAddress.Address, nil
}

func (c *ComputeUtil) region() string {
	return c.zone[:len(c.zone)-2]
}

func (c *ComputeUtil) firewallRule() (*raw.Firewall, error) {
	return c.service.Firewalls.Get(c.project, firewallRule).Do()
}

func (c *ComputeUtil) createFirewallRule() error {
	log.Infof("Creating firewall rule.")
	allowed := []*raw.FirewallAllowed{

		{
			IPProtocol: "tcp",
			Ports: []string{
				port,
			},
		},
	}

	if c.SwarmMaster {
		u, err := url.Parse(c.SwarmHost)
		if err != nil {
			return fmt.Errorf("error authorizing port for swarm: %s", err)
		}

		parts := strings.Split(u.Host, ":")
		swarmPort := parts[1]
		allowed = append(allowed, &raw.FirewallAllowed{
			IPProtocol: "tcp",
			Ports: []string{
				swarmPort,
			},
		})
	}
	rule := &raw.Firewall{
		Allowed: allowed,
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

	tags := append(d.Tags, firewallTargetTag)

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
			Items: tags,
		},
		ServiceAccounts: []*raw.ServiceAccount{
			{
				Email:  "default",
				Scopes: strings.Split(d.Scopes, ","),
			},
		},
		Scheduling: &raw.Scheduling{
			Preemptible: c.preemptible,
		},
	}

	if c.address != "" {
		staticAddress, err := c.staticAddress()
		if err != nil {
			return err
		}

		instance.NetworkInterfaces[0].AccessConfigs[0].NatIP = staticAddress
	}

	disk, err := c.disk()
	if disk == nil || err != nil {
		instance.Disks[0].InitializeParams = &raw.AttachedDiskInitializeParams{
			DiskName:    c.diskName(),
			SourceImage: imageName,
			// The maximum supported disk size is 1000GB, the cast should be fine.
			DiskSizeGb: int64(d.DiskSize),
			DiskType:   c.diskType(),
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

	// Update the SSH Key
	sshKey, err := ioutil.ReadFile(d.GetSSHKeyPath() + ".pub")
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

// stopInstance stops the instance.
func (c *ComputeUtil) stopInstance() error {
	log.Infof("Stopping instance.")
	op, err := c.service.Instances.Stop(c.project, c.zone, c.instanceName).Do()
	if err != nil {
		return err
	}

	log.Infof("Waiting for instance to stop.")
	return c.waitForRegionalOp(op.Name)
}

// startInstance starts the instance.
func (c *ComputeUtil) startInstance() error {
	log.Infof("Starting instance.")
	op, err := c.service.Instances.Start(c.project, c.zone, c.instanceName).Do()
	if err != nil {
		return err
	}

	log.Infof("Waiting for instance to start.")
	return c.waitForRegionalOp(op.Name)
}

func (c *ComputeUtil) executeCommands(commands []string, ip, sshKeyPath string) error {
	for _, command := range commands {
		auth := &ssh.Auth{
			Keys: []string{sshKeyPath},
		}

		client, err := ssh.NewClient(c.userName, ip, 22, auth)
		if err != nil {
			return err
		}

		if _, err := client.Output(command); err != nil {
			return err
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

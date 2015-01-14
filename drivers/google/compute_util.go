package google

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/drivers"
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
	apiURL             = "https://www.googleapis.com/compute/v1/projects/"
	imageName          = "https://www.googleapis.com/compute/v1/projects/google-containers/global/images/container-vm-v20141016"
	firewallRule       = "docker-machines"
	port               = "2376"
	firewallTargetTag  = "docker-machine"
	dockerStartCommand = "sudo service docker start"
	dockerStopCommand  = "sudo service docker stop"
)

var (
	dockerUpgradeScriptTemplate = template.Must(template.New("upgrade-docker-script").Parse(
		`sudo mkdir -p /.docker/authorized-keys.d/
sudo chown -R {{ .UserName }} /.docker
while [ -e /var/run/docker.pid ]; do sleep 1; done
sudo sed -i 's@DOCKER_OPTS=.*@DOCKER_OPTS=\"--auth=identity -H unix://var/run/docker.sock -H 0.0.0.0:2376\"@g' /etc/default/docker
sudo wget https://bfirsh.s3.amazonaws.com/docker/docker-1.3.1-dev-identity-auth -O /usr/bin/docker && sudo chmod +x /usr/bin/docker
`))
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

	return c.updateDocker(d)
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

// updateDocker updates the docker daemon to the latest version.
func (c *ComputeUtil) updateDocker(d *Driver) error {
	log.Infof("Updating docker.")
	ip, err := d.GetIP()
	if err != nil {
		return fmt.Errorf("error retrieving ip: %v", err)
	}
	if c.executeCommands([]string{dockerStopCommand}, ip, d.sshKeyPath); err != nil {
		return err
	}
	var scriptBuf bytes.Buffer

	if err := dockerUpgradeScriptTemplate.Execute(&scriptBuf, d); err != nil {
		return fmt.Errorf("error expanding upgrade script template: %v", err)
	}
	commands := strings.Split(scriptBuf.String(), "\n")
	if err := c.executeCommands(commands, ip, d.sshKeyPath); err != nil {
		return err
	}
	if err := drivers.AddPublicKeyToAuthorizedHosts(d, "/.docker/authorized-keys.d"); err != nil {
		return err
	}
	return c.executeCommands([]string{dockerStartCommand}, ip, d.sshKeyPath)
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

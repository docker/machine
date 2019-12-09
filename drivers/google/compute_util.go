package google

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/docker/machine/drivers/driverutil"
	"github.com/docker/machine/libmachine/log"
	raw "google.golang.org/api/compute/v1"

	"errors"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/googleapi"
)

// ComputeUtil is used to wrap the raw GCE API code and store common parameters.
type ComputeUtil struct {
	zone              string
	instanceName      string
	userName          string
	project           string
	diskTypeURL       string
	address           string
	networkProject    string
	network           string
	subnetwork        string
	preemptible       bool
	useInternalIP     bool
	useInternalIPOnly bool
	service           *raw.Service
	zoneURL           string
	globalURL         string
	SwarmMaster       bool
	SwarmHost         string
	openPorts         []string
}

const (
	apiURL            = "https://www.googleapis.com/compute/v1/projects/"
	firewallRule      = "docker-machines"
	dockerPort        = "2376"
	firewallTargetTag = "docker-machine"
)

// NewComputeUtil creates and initializes a ComputeUtil.
func newComputeUtil(driver *Driver) (*ComputeUtil, error) {
	client, err := google.DefaultClient(oauth2.NoContext, raw.ComputeScope)
	if err != nil {
		return nil, err
	}

	service, err := raw.New(client)
	if err != nil {
		return nil, err
	}

	var networkProject string
	if strings.Contains(driver.Network, "/projects/") {
		var splittedElements = strings.Split(driver.Network, "/")
		for i, element := range splittedElements {
			if element == "projects" {
				networkProject = splittedElements[i+1]
				break
			}
		}
	} else {
		networkProject = driver.Project
	}

	return &ComputeUtil{
		zone:              driver.Zone,
		instanceName:      driver.MachineName,
		userName:          driver.SSHUser,
		project:           driver.Project,
		diskTypeURL:       driver.DiskType,
		address:           driver.Address,
		networkProject:    networkProject,
		network:           driver.Network,
		subnetwork:        driver.Subnetwork,
		preemptible:       driver.Preemptible,
		useInternalIP:     driver.UseInternalIP,
		useInternalIPOnly: driver.UseInternalIPOnly,
		service:           service,
		zoneURL:           apiURL + driver.Project + "/zones/" + driver.Zone,
		globalURL:         apiURL + driver.Project + "/global",
		SwarmMaster:       driver.SwarmMaster,
		SwarmHost:         driver.SwarmHost,
		openPorts:         driver.OpenPorts,
	}, nil
}

func (c *ComputeUtil) diskName() string {
	return c.instanceName + "-disk"
}

func (c *ComputeUtil) diskType() string {
	return apiURL + c.project + "/zones/" + c.zone + "/diskTypes/" + c.diskTypeURL
}

// disk returns the persistent disk attached to the vm.
func (c *ComputeUtil) disk() (*raw.Disk, error) {
	return c.service.Disks.Get(c.project, c.zone, c.diskName()).Do()
}

// deleteDisk deletes the persistent disk.
func (c *ComputeUtil) deleteDisk() error {
	disk, _ := c.disk()
	if disk == nil {
		return nil
	}

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
	return c.service.Firewalls.Get(c.networkProject, firewallRule).Do()
}

func missingOpenedPorts(rule *raw.Firewall, ports []string) map[string][]string {
	missing := map[string][]string{}
	opened := map[string]bool{}

	for _, allowed := range rule.Allowed {
		for _, allowedPort := range allowed.Ports {
			opened[allowedPort+"/"+allowed.IPProtocol] = true
		}
	}

	for _, p := range ports {
		port, proto := driverutil.SplitPortProto(p)
		if !opened[port+"/"+proto] {
			missing[proto] = append(missing[proto], port)
		}
	}

	return missing
}

func (c *ComputeUtil) portsUsed() ([]string, error) {
	ports := []string{dockerPort + "/tcp"}

	if c.SwarmMaster {
		u, err := url.Parse(c.SwarmHost)
		if err != nil {
			return nil, fmt.Errorf("error authorizing port for swarm: %s", err)
		}

		swarmPort := strings.Split(u.Host, ":")[1]
		ports = append(ports, swarmPort+"/tcp")
	}
	for _, p := range c.openPorts {
		port, proto := driverutil.SplitPortProto(p)
		ports = append(ports, port+"/"+proto)
	}

	return ports, nil
}

// openFirewallPorts configures the firewall to open docker and swarm ports.
func (c *ComputeUtil) openFirewallPorts(d *Driver) error {
	log.Infof("Opening firewall ports")

	create := false
	rule, err := c.firewallRule()
	if err != nil {
		return err
	}
	if rule == nil {
		create = true
		var net string
		if strings.Contains(d.Network, "/networks/") {
			net = d.Network
		} else {
			net = c.globalURL + "/networks/" + d.Network
		}
		rule = &raw.Firewall{
			Name:         firewallRule,
			Allowed:      []*raw.FirewallAllowed{},
			SourceRanges: []string{"0.0.0.0/0"},
			TargetTags:   []string{firewallTargetTag},
			Network:      net,
		}
	}

	portsUsed, err := c.portsUsed()
	if err != nil {
		return err
	}

	missingPorts := missingOpenedPorts(rule, portsUsed)
	if len(missingPorts) == 0 {
		return nil
	}
	for proto, ports := range missingPorts {
		rule.Allowed = append(rule.Allowed, &raw.FirewallAllowed{
			IPProtocol: proto,
			Ports:      ports,
		})
	}

	var op *raw.Operation
	if create {
		op, err = c.service.Firewalls.Insert(c.networkProject, rule).Do()
	} else {
		op, err = c.service.Firewalls.Update(c.networkProject, firewallRule, rule).Do()
	}

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
	log.Infof("Creating instance")

	var net string
	if strings.Contains(d.Network, "/networks/") {
		net = d.Network
	} else {
		net = c.globalURL + "/networks/" + d.Network
	}

	instance := &raw.Instance{
		Name:        c.instanceName,
		Description: "docker host vm",
		MachineType: c.zoneURL + "/machineTypes/" + d.MachineType,
		Disks: []*raw.AttachedDisk{
			{
				Boot:       true,
				AutoDelete: true,
				Type:       "PERSISTENT",
				Mode:       "READ_WRITE",
			},
		},
		NetworkInterfaces: []*raw.NetworkInterface{
			{
				Network: net,
			},
		},
		Tags: &raw.Tags{
			Items: parseTags(d),
		},
		ServiceAccounts: []*raw.ServiceAccount{
			{
				Email:  d.ServiceAccount,
				Scopes: strings.Split(d.Scopes, ","),
			},
		},
		Scheduling: &raw.Scheduling{
			Preemptible: c.preemptible,
		},
	}

	if strings.Contains(c.subnetwork, "/subnetworks/") {
		instance.NetworkInterfaces[0].Subnetwork = c.subnetwork
	} else if c.subnetwork != "" {
		instance.NetworkInterfaces[0].Subnetwork = "projects/" + c.networkProject + "/regions/" + c.region() + "/subnetworks/" + c.subnetwork
	}

	if !c.useInternalIPOnly {
		cfg := &raw.AccessConfig{
			Type: "ONE_TO_ONE_NAT",
		}
		instance.NetworkInterfaces[0].AccessConfigs = append(instance.NetworkInterfaces[0].AccessConfigs, cfg)
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
			SourceImage: "https://www.googleapis.com/compute/v1/projects/" + d.MachineImage,
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

	log.Infof("Waiting for Instance")
	if err = c.waitForRegionalOp(op.Name); err != nil {
		return err
	}

	instance, err = c.instance()
	if err != nil {
		return err
	}

	return c.uploadSSHKey(instance, d.GetSSHKeyPath())
}

// configureInstance configures an existing instance for use with Docker Machine.
func (c *ComputeUtil) configureInstance(d *Driver) error {
	log.Infof("Configuring instance")

	instance, err := c.instance()
	if err != nil {
		return err
	}

	if err := c.addFirewallTag(instance); err != nil {
		return err
	}

	return c.uploadSSHKey(instance, d.GetSSHKeyPath())
}

// addFirewallTag adds a tag to the instance to match the firewall rule.
func (c *ComputeUtil) addFirewallTag(instance *raw.Instance) error {
	log.Infof("Adding tag for the firewall rule")

	tags := instance.Tags
	for _, tag := range tags.Items {
		if tag == firewallTargetTag {
			return nil
		}
	}

	tags.Items = append(tags.Items, firewallTargetTag)

	op, err := c.service.Instances.SetTags(c.project, c.zone, instance.Name, tags).Do()
	if err != nil {
		return err
	}

	return c.waitForRegionalOp(op.Name)
}

// uploadSSHKey updates the instance metadata with the given ssh key.
func (c *ComputeUtil) uploadSSHKey(instance *raw.Instance, sshKeyPath string) error {
	log.Infof("Uploading SSH Key")

	sshKey, err := ioutil.ReadFile(sshKeyPath + ".pub")
	if err != nil {
		return err
	}

	metaDataValue := fmt.Sprintf("%s:%s %s\n", c.userName, strings.TrimSpace(string(sshKey)), c.userName)

	op, err := c.service.Instances.SetMetadata(c.project, c.zone, c.instanceName, &raw.Metadata{
		Fingerprint: instance.Metadata.Fingerprint,
		Items: []*raw.MetadataItems{
			{
				Key:   "sshKeys",
				Value: &metaDataValue,
			},
		},
	}).Do()

	return c.waitForRegionalOp(op.Name)
}

// parseTags computes the tags for the instance.
func parseTags(d *Driver) []string {
	tags := []string{firewallTargetTag}

	if d.Tags != "" {
		tags = append(tags, strings.Split(d.Tags, ",")...)
	}

	return tags
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
	op, err := c.service.Instances.Stop(c.project, c.zone, c.instanceName).Do()
	if err != nil {
		return err
	}

	log.Infof("Waiting for instance to stop.")
	return c.waitForRegionalOp(op.Name)
}

// startInstance starts the instance.
func (c *ComputeUtil) startInstance() error {
	op, err := c.service.Instances.Start(c.project, c.zone, c.instanceName).Do()
	if err != nil {
		return err
	}

	log.Infof("Waiting for instance to start.")
	return c.waitForRegionalOp(op.Name)
}

// waitForOp waits for the operation to finish.
func (c *ComputeUtil) waitForOp(opGetter func() (*raw.Operation, error)) error {
	for {
		op, err := opGetter()
		if err != nil {
			return err
		}

		log.Debugf("Operation %q status: %s", op.Name, op.Status)
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

// waitForRegionalOp waits for the regional operation to finish.
func (c *ComputeUtil) waitForRegionalOp(name string) error {
	return c.waitForOp(func() (*raw.Operation, error) {
		return c.service.ZoneOperations.Get(c.project, c.zone, name).Do()
	})
}

// waitForGlobalOp waits for the global operation to finish.
func (c *ComputeUtil) waitForGlobalOp(name string) error {
	return c.waitForOp(func() (*raw.Operation, error) {
		return c.service.GlobalOperations.Get(c.project, name).Do()
	})
}

// ip retrieves and returns the external IP address of the instance.
func (c *ComputeUtil) ip() (string, error) {
	instance, err := c.service.Instances.Get(c.project, c.zone, c.instanceName).Do()
	if err != nil {
		return "", unwrapGoogleError(err)
	}

	nic := instance.NetworkInterfaces[0]
	if c.useInternalIP {
		return nic.NetworkIP, nil
	}
	return nic.AccessConfigs[0].NatIP, nil
}

func unwrapGoogleError(err error) error {
	if googleErr, ok := err.(*googleapi.Error); ok {
		return errors.New(googleErr.Message)
	}

	return err
}

func isNotFound(err error) bool {
	googleErr, ok := err.(*googleapi.Error)
	if !ok {
		return false
	}

	if googleErr.Code == http.StatusNotFound {
		return true
	}

	return false
}

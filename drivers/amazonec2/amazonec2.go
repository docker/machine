package amazonec2

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/drivers/amazonec2/amz"
	"github.com/docker/machine/provider"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
)

const (
	driverName               = "amazonec2"
	defaultRegion            = "us-east-1"
	defaultInstanceType      = "t2.micro"
	defaultRootSize          = 16
	ipRange                  = "0.0.0.0/0"
	machineSecurityGroupName = "docker-machine"
)

var (
	dockerPort = 2376
	swarmPort  = 3376
)

type Driver struct {
	Id                  string
	AccessKey           string
	SecretKey           string
	SessionToken        string
	Region              string
	AMI                 string
	SSHKeyID            int
	SSHUser             string
	SSHPort             int
	KeyName             string
	InstanceId          string
	InstanceType        string
	IPAddress           string
	PrivateIPAddress    string
	MachineName         string
	SecurityGroupId     string
	SecurityGroupName   string
	ReservationId       string
	RootSize            int64
	IamInstanceProfile  string
	VpcId               string
	SubnetId            string
	Zone                string
	CaCertPath          string
	PrivateKeyPath      string
	SwarmMaster         bool
	SwarmHost           string
	SwarmDiscovery      string
	storePath           string
	keyPath             string
	RequestSpotInstance bool
	SpotPrice           string
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
			Name:   "amazonec2-access-key",
			Usage:  "AWS Access Key",
			Value:  "",
			EnvVar: "AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "amazonec2-secret-key",
			Usage:  "AWS Secret Key",
			Value:  "",
			EnvVar: "AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "amazonec2-session-token",
			Usage:  "AWS Session Token",
			Value:  "",
			EnvVar: "AWS_SESSION_TOKEN",
		},
		cli.StringFlag{
			Name:   "amazonec2-ami",
			Usage:  "AWS machine image",
			EnvVar: "AWS_AMI",
		},
		cli.StringFlag{
			Name:   "amazonec2-region",
			Usage:  "AWS region",
			Value:  defaultRegion,
			EnvVar: "AWS_DEFAULT_REGION",
		},
		cli.StringFlag{
			Name:   "amazonec2-vpc-id",
			Usage:  "AWS VPC id",
			Value:  "",
			EnvVar: "AWS_VPC_ID",
		},
		cli.StringFlag{
			Name:   "amazonec2-zone",
			Usage:  "AWS zone for instance (i.e. a,b,c,d,e)",
			Value:  "a",
			EnvVar: "AWS_ZONE",
		},
		cli.StringFlag{
			Name:   "amazonec2-subnet-id",
			Usage:  "AWS VPC subnet id",
			Value:  "",
			EnvVar: "AWS_SUBNET_ID",
		},
		cli.StringFlag{
			Name:   "amazonec2-security-group",
			Usage:  "AWS VPC security group",
			Value:  "docker-machine",
			EnvVar: "AWS_SECURITY_GROUP",
		},
		cli.StringFlag{
			Name:   "amazonec2-instance-type",
			Usage:  "AWS instance type",
			Value:  defaultInstanceType,
			EnvVar: "AWS_INSTANCE_TYPE",
		},
		cli.IntFlag{
			Name:   "amazonec2-root-size",
			Usage:  "AWS root disk size (in GB)",
			Value:  defaultRootSize,
			EnvVar: "AWS_ROOT_SIZE",
		},
		cli.StringFlag{
			Name:  "amazonec2-iam-instance-profile",
			Usage: "AWS IAM Instance Profile",
		},
		cli.StringFlag{
			Name:   "amazonec2-ssh-user",
			Usage:  "set the name of the ssh user",
			Value:  "ubuntu",
			EnvVar: "AWS_SSH_USER",
		},
		cli.BoolFlag{
			Name:  "amazonec2-request-spot-instance",
			Usage: "Set this flag to request spot instance",
		},
		cli.StringFlag{
			Name:  "amazonec2-spot-price",
			Usage: "AWS spot instance bid price (in dollar)",
			Value: "0.50",
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	id := generateId()
	return &Driver{
		Id:             id,
		MachineName:    machineName,
		storePath:      storePath,
		CaCertPath:     caCert,
		PrivateKeyPath: privateKey,
	}, nil
}

func (d *Driver) GetProviderType() provider.ProviderType {
	return provider.Remote
}

func (d *Driver) AuthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) DeauthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	region, err := validateAwsRegion(flags.String("amazonec2-region"))
	if err != nil {
		return err
	}

	image := flags.String("amazonec2-ami")
	if len(image) == 0 {
		image = regionDetails[region].AmiId
	}

	d.AccessKey = flags.String("amazonec2-access-key")
	d.SecretKey = flags.String("amazonec2-secret-key")
	d.SessionToken = flags.String("amazonec2-session-token")
	d.Region = region
	d.AMI = image
	d.RequestSpotInstance = flags.Bool("amazonec2-request-spot-instance")
	d.SpotPrice = flags.String("amazonec2-spot-price")
	d.InstanceType = flags.String("amazonec2-instance-type")
	d.VpcId = flags.String("amazonec2-vpc-id")
	d.SubnetId = flags.String("amazonec2-subnet-id")
	d.SecurityGroupName = flags.String("amazonec2-security-group")
	zone := flags.String("amazonec2-zone")
	d.Zone = zone[:]
	d.RootSize = int64(flags.Int("amazonec2-root-size"))
	d.IamInstanceProfile = flags.String("amazonec2-iam-instance-profile")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = flags.String("amazonec2-ssh-user")
	d.SSHPort = 22

	if d.AccessKey == "" {
		return fmt.Errorf("amazonec2 driver requires the --amazonec2-access-key option")
	}

	if d.SecretKey == "" {
		return fmt.Errorf("amazonec2 driver requires the --amazonec2-secret-key option")
	}

	if d.SubnetId == "" && d.VpcId == "" {
		return fmt.Errorf("amazonec2 driver requires either the --amazonec2-subnet-id or --amazonec2-vpc-id option")
	}

	if d.isSwarmMaster() {
		u, err := url.Parse(d.SwarmHost)
		if err != nil {
			return fmt.Errorf("error parsing swarm host: %s", err)
		}

		parts := strings.Split(u.Host, ":")
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			return err
		}

		swarmPort = port
	}

	return nil
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) DriverName() string {
	return driverName
}

func (d *Driver) checkPrereqs() error {
	// check for existing keypair
	key, err := d.getClient().GetKeyPair(d.MachineName)
	if err != nil {
		return err
	}

	if key != nil {
		return fmt.Errorf("There is already a keypair with the name %s.  Please either remove that keypair or use a different machine name.", d.MachineName)
	}

	regionZone := d.Region + d.Zone
	if d.SubnetId == "" {
		filters := []amz.Filter{
			{
				Name:  "availabilityZone",
				Value: regionZone,
			},
			{
				Name:  "vpc-id",
				Value: d.VpcId,
			},
		}

		subnets, err := d.getClient().GetSubnets(filters)
		if err != nil {
			return err
		}

		if len(subnets) == 0 {
			return fmt.Errorf("unable to find a subnet in the zone: %s", regionZone)
		}

		d.SubnetId = subnets[0].SubnetId

		// try to find default
		if len(subnets) > 1 {
			for _, subnet := range subnets {
				if subnet.DefaultForAz {
					d.SubnetId = subnet.SubnetId
					break
				}
			}
		}
	}

	return nil
}

func (d *Driver) PreCreateCheck() error {
	return d.checkPrereqs()
}

func (d *Driver) instanceIpAvailable() bool {
	ip, err := d.GetIP()
	if err != nil {
		log.Debug(err)
	}
	if ip != "" {
		d.IPAddress = ip
		log.Debugf("Got the IP Address, it's %q", d.IPAddress)
		return true
	}
	return false
}

func (d *Driver) Create() error {
	if err := d.checkPrereqs(); err != nil {
		return err
	}

	log.Infof("Launching instance...")

	if err := d.createKeyPair(); err != nil {
		return fmt.Errorf("unable to create key pair: %s", err)
	}

	if err := d.configureSecurityGroup(d.SecurityGroupName); err != nil {
		return err
	}

	bdm := &amz.BlockDeviceMapping{
		DeviceName:          "/dev/sda1",
		VolumeSize:          d.RootSize,
		DeleteOnTermination: true,
		VolumeType:          "gp2",
	}

	log.Debugf("launching instance in subnet %s", d.SubnetId)
	var instance amz.EC2Instance
	if d.RequestSpotInstance {
		spotInstanceRequestId, err := d.getClient().RequestSpotInstances(d.AMI, d.InstanceType, d.Zone, 1, d.SecurityGroupId, d.KeyName, d.SubnetId, bdm, d.IamInstanceProfile, d.SpotPrice)
		if err != nil {
			return fmt.Errorf("Error request spot instance: %s", err)
		}
		var instanceId string
		var spotInstanceRequestStatus string
		log.Info("Waiting for spot instance...")
		// check until fulfilled
		for instanceId == "" {
			time.Sleep(time.Second * 5)
			spotInstanceRequestStatus, instanceId, err = d.getClient().DescribeSpotInstanceRequests(spotInstanceRequestId)
			if err != nil {
				return fmt.Errorf("Error describe spot instance request: %s", err)
			}
			log.Debugf("spot instance request status: %s", spotInstanceRequestStatus)
		}
		instance, err = d.getClient().GetInstance(instanceId)
		if err != nil {
			return fmt.Errorf("Error get instance: %s", err)
		}
	} else {
		inst, err := d.getClient().RunInstance(d.AMI, d.InstanceType, d.Zone, 1, 1, d.SecurityGroupId, d.KeyName, d.SubnetId, bdm, d.IamInstanceProfile)
		if err != nil {
			return fmt.Errorf("Error launching instance: %s", err)
		}
		instance = inst
	}

	d.InstanceId = instance.InstanceId

	log.Debug("waiting for ip address to become available")
	if err := utils.WaitFor(d.instanceIpAvailable); err != nil {
		return err
	}

	if len(instance.NetworkInterfaceSet) > 0 {
		d.PrivateIPAddress = instance.NetworkInterfaceSet[0].PrivateIpAddress
	}

	d.waitForInstance()

	log.Debugf("created instance ID %s, IP address %s, Private IP address %s",
		d.InstanceId,
		d.IPAddress,
		d.PrivateIPAddress,
	)

	log.Debug("Settings tags for instance")
	tags := map[string]string{
		"Name": d.MachineName,
	}

	if err := d.getClient().CreateTags(d.InstanceId, tags); err != nil {
		return err
	}

	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:%d", ip, dockerPort), nil
}

func (d *Driver) GetIP() (string, error) {
	inst, err := d.getInstance()
	if err != nil {
		return "", err
	}

	return inst.IpAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	inst, err := d.getInstance()
	if err != nil {
		return state.Error, err
	}
	switch inst.InstanceState.Name {
	case "pending":
		return state.Starting, nil
	case "running":
		return state.Running, nil
	case "stopping":
		return state.Stopping, nil
	case "shutting-down":
		return state.Stopping, nil
	case "stopped":
		return state.Stopped, nil
	default:
		return state.Error, nil
	}
	return state.None, nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	// TODO: use @nathanleclaire retry func here (ehazlett)
	return d.GetIP()
}

func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = 22
	}

	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "ubuntu"
	}

	return d.SSHUser
}

func (d *Driver) Start() error {
	if err := d.getClient().StartInstance(d.InstanceId); err != nil {
		return err
	}

	if err := d.waitForInstance(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Stop() error {
	if err := d.getClient().StopInstance(d.InstanceId, false); err != nil {
		return err
	}
	return nil
}

func (d *Driver) Remove() error {

	if err := d.terminate(); err != nil {
		return fmt.Errorf("unable to terminate instance: %s", err)
	}

	// remove keypair
	if err := d.deleteKeyPair(); err != nil {
		return fmt.Errorf("unable to remove key pair: %s", err)
	}

	return nil
}

func (d *Driver) Restart() error {
	if err := d.getClient().RestartInstance(d.InstanceId); err != nil {
		return fmt.Errorf("unable to restart instance: %s", err)
	}
	return nil
}

func (d *Driver) Kill() error {
	if err := d.getClient().StopInstance(d.InstanceId, true); err != nil {
		return err
	}
	return nil
}

func (d *Driver) getClient() *amz.EC2 {
	auth := amz.GetAuth(d.AccessKey, d.SecretKey, d.SessionToken)
	return amz.NewEC2(auth, d.Region)
}

func (d *Driver) GetSSHKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) getInstance() (*amz.EC2Instance, error) {
	instance, err := d.getClient().GetInstance(d.InstanceId)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (d *Driver) instanceIsRunning() bool {
	st, err := d.GetState()
	if err != nil {
		log.Debug(err)
	}
	if st == state.Running {
		return true
	}
	return false
}

func (d *Driver) waitForInstance() error {
	if err := utils.WaitFor(d.instanceIsRunning); err != nil {
		return err
	}

	return nil
}

func (d *Driver) createKeyPair() error {

	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	publicKey, err := ioutil.ReadFile(d.GetSSHKeyPath() + ".pub")
	if err != nil {
		return err
	}

	keyName := d.MachineName

	log.Debugf("creating key pair: %s", keyName)

	if err := d.getClient().ImportKeyPair(keyName, string(publicKey)); err != nil {
		return err
	}

	d.KeyName = keyName
	return nil
}

func (d *Driver) terminate() error {
	if d.InstanceId == "" {
		return fmt.Errorf("unknown instance")
	}

	log.Debugf("terminating instance: %s", d.InstanceId)
	if err := d.getClient().TerminateInstance(d.InstanceId); err != nil {
		return fmt.Errorf("unable to terminate instance: %s", err)
	}

	return nil
}

func (d *Driver) isSwarmMaster() bool {
	return d.SwarmMaster
}

func (d *Driver) securityGroupAvailableFunc(id string) func() bool {
	return func() bool {
		_, err := d.getClient().GetSecurityGroupById(id)
		if err == nil {
			return true
		}
		log.Debug(err)
		return false
	}
}

func (d *Driver) configureSecurityGroup(groupName string) error {
	log.Debugf("configuring security group in %s", d.VpcId)

	var securityGroup *amz.SecurityGroup

	groups, err := d.getClient().GetSecurityGroups()
	if err != nil {
		return err
	}

	for _, grp := range groups {
		if grp.GroupName == groupName {
			log.Debugf("found existing security group (%s) in %s", groupName, d.VpcId)
			securityGroup = &grp
			break
		}
	}

	// if not found, create
	if securityGroup == nil {
		log.Debugf("creating security group (%s) in %s", groupName, d.VpcId)
		group, err := d.getClient().CreateSecurityGroup(groupName, "Docker Machine", d.VpcId)
		if err != nil {
			return err
		}
		securityGroup = group
		// wait until created (dat eventual consistency)
		log.Debugf("waiting for group (%s) to become available", group.GroupId)
		if err := utils.WaitFor(d.securityGroupAvailableFunc(group.GroupId)); err != nil {
			return err
		}
	}

	d.SecurityGroupId = securityGroup.GroupId

	perms := d.configureSecurityGroupPermissions(securityGroup)

	if len(perms) != 0 {
		log.Debugf("authorizing group %s with permissions: %v", securityGroup.GroupName, perms)
		if err := d.getClient().AuthorizeSecurityGroup(d.SecurityGroupId, perms); err != nil {
			return err
		}

	}

	return nil
}

func (d *Driver) configureSecurityGroupPermissions(group *amz.SecurityGroup) []amz.IpPermission {
	hasSshPort := false
	hasDockerPort := false
	hasSwarmPort := false
	for _, p := range group.IpPermissions {
		switch p.FromPort {
		case 22:
			hasSshPort = true
		case dockerPort:
			hasDockerPort = true
		case swarmPort:
			hasSwarmPort = true
		}
	}

	perms := []amz.IpPermission{}

	if !hasSshPort {
		perms = append(perms, amz.IpPermission{
			IpProtocol: "tcp",
			FromPort:   22,
			ToPort:     22,
			IpRange:    ipRange,
		})
	}

	if !hasDockerPort {
		perms = append(perms, amz.IpPermission{
			IpProtocol: "tcp",
			FromPort:   dockerPort,
			ToPort:     dockerPort,
			IpRange:    ipRange,
		})
	}

	if !hasSwarmPort && d.SwarmMaster {
		perms = append(perms, amz.IpPermission{
			IpProtocol: "tcp",
			FromPort:   swarmPort,
			ToPort:     swarmPort,
			IpRange:    ipRange,
		})
	}

	log.Debugf("configuring security group authorization for %s", ipRange)

	return perms
}

func (d *Driver) deleteSecurityGroup() error {
	log.Debugf("deleting security group %s", d.SecurityGroupId)

	if err := d.getClient().DeleteSecurityGroup(d.SecurityGroupId); err != nil {
		return err
	}

	return nil
}

func (d *Driver) deleteKeyPair() error {
	log.Debugf("deleting key pair: %s", d.KeyName)

	if err := d.getClient().DeleteKeyPair(d.KeyName); err != nil {
		return err
	}

	return nil
}

func generateId() string {
	rb := make([]byte, 10)
	_, err := rand.Read(rb)
	if err != nil {
		log.Fatalf("unable to generate id: %s", err)
	}

	h := md5.New()
	io.WriteString(h, string(rb))
	return fmt.Sprintf("%x", h.Sum(nil))
}

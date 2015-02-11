package amazonec2

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"path"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/drivers/amazonec2/amz"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

const (
	driverName               = "amazonec2"
	defaultRegion            = "us-east-1"
	defaultInstanceType      = "t2.micro"
	defaultRootSize          = 16
	ipRange                  = "0.0.0.0/0"
	dockerConfigDir          = "/etc/docker"
	machineSecurityGroupName = "docker-machine"
	dockerPort               = 2376
)

type Driver struct {
	Id                string
	AccessKey         string
	SecretKey         string
	SessionToken      string
	Region            string
	AMI               string
	SSHKeyID          int
	KeyName           string
	InstanceId        string
	InstanceType      string
	IPAddress         string
	MachineName       string
	SecurityGroupName string
	SecurityGroupId   string
	ReservationId     string
	RootSize          int64
	VpcId             string
	SubnetId          string
	Zone              string
	CaCertPath        string
	PrivateKeyPath    string
	storePath         string
	keyPath           string
}

type CreateFlags struct {
	AccessKey    *string
	SecretKey    *string
	Region       *string
	AMI          *string
	InstanceType *string
	SubnetId     *string
	RootSize     *int64
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
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	id := generateId()
	return &Driver{Id: id, MachineName: machineName, storePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey}, nil
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	region, err := validateAwsRegion(flags.String("amazonec2-region"))
	if err != nil {
		return nil
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
	d.InstanceType = flags.String("amazonec2-instance-type")
	d.VpcId = flags.String("amazonec2-vpc-id")
	d.SubnetId = flags.String("amazonec2-subnet-id")
	d.SecurityGroupName = flags.String("amazonec2-security-group")
	zone := flags.String("amazonec2-zone")
	d.Zone = zone[:]
	d.RootSize = int64(flags.Int("amazonec2-root-size"))

	if d.AccessKey == "" {
		return fmt.Errorf("amazonec2 driver requires the --amazonec2-access-key option")
	}

	if d.SecretKey == "" {
		return fmt.Errorf("amazonec2 driver requires the --amazonec2-secret-key option")
	}

	if d.SubnetId == "" && d.VpcId == "" {
		return fmt.Errorf("amazonec2 driver requires either the --amazonec2-subnet-id or --amazonec2-vpc-id option")
	}

	return nil
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
	return nil
}

func (d *Driver) PreCreateCheck() error {
	return d.checkPrereqs()
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

	// get the subnet id
	regionZone := d.Region + d.Zone
	subnetId := d.SubnetId

	if d.SubnetId == "" {
		subnets, err := d.getClient().GetSubnets()
		if err != nil {
			return err
		}

		for _, s := range subnets {
			if s.AvailabilityZone == regionZone {
				subnetId = s.SubnetId
				break
			}
		}

	}

	if subnetId == "" {
		return fmt.Errorf("unable to find a subnet in the zone: %s", regionZone)
	}

	log.Debugf("launching instance in subnet %s", subnetId)
	instance, err := d.getClient().RunInstance(d.AMI, d.InstanceType, d.Zone, 1, 1, d.SecurityGroupId, d.KeyName, subnetId, bdm)

	if err != nil {
		return fmt.Errorf("Error launching instance: %s", err)
	}

	d.InstanceId = instance.InstanceId

	d.waitForInstance()

	log.Debugf("created instance ID %s, IP address %s",
		d.InstanceId,
		d.IPAddress)

	log.Infof("Waiting for SSH on %s:%d", d.IPAddress, 22)

	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", d.IPAddress, 22)); err != nil {
		return err
	}

	log.Debug("Settings tags for instance")
	tags := map[string]string{
		"Name": d.MachineName,
	}

	if err = d.getClient().CreateTags(d.InstanceId, tags); err != nil {
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

	cmd, err = d.GetSSHCommand("if [ ! -e /usr/bin/docker ]; then curl -sL https://get.docker.com | sh -; fi")
	if err != nil {
		return err

	}
	if err := cmd.Run(); err != nil {
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
	d.IPAddress = inst.IpAddress
	return d.IPAddress, nil
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
	}
	return state.None, nil
}

func (d *Driver) Start() error {
	if err := d.getClient().StartInstance(d.InstanceId); err != nil {
		return err
	}

	if err := d.waitForInstance(); err != nil {
		return err
	}

	if err := d.updateDriver(); err != nil {
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

func (d *Driver) StartDocker() error {
	log.Debug("Starting Docker...")

	cmd, err := d.GetSSHCommand("sudo service docker start")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) StopDocker() error {
	log.Debug("Stopping Docker...")

	cmd, err := d.GetSSHCommand("sudo service docker stop")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) GetDockerConfigDir() string {
	return dockerConfigDir
}

func (d *Driver) Upgrade() error {
	log.Debugf("Upgrading Docker")

	cmd, err := d.GetSSHCommand("sudo apt-get update && sudo apt-get install --upgrade lxc-docker")
	if err != nil {
		return err

	}
	if err := cmd.Run(); err != nil {
		return err

	}

	return cmd.Run()
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return ssh.GetSSHCommand(d.IPAddress, 22, "ubuntu", d.sshKeyPath(), args...), nil
}

func (d *Driver) getClient() *amz.EC2 {
	auth := amz.GetAuth(d.AccessKey, d.SecretKey, d.SessionToken)
	return amz.NewEC2(auth, d.Region)
}

func (d *Driver) sshKeyPath() string {
	return path.Join(d.storePath, "id_rsa")
}

func (d *Driver) updateDriver() error {
	inst, err := d.getInstance()
	if err != nil {
		return err
	}
	// wait for ipaddress
	for {
		i, err := d.getInstance()
		if err != nil {
			return err
		}
		if i.IpAddress == "" {
			time.Sleep(1 * time.Second)
			continue
		}

		d.InstanceId = inst.InstanceId
		d.IPAddress = inst.IpAddress
		break
	}
	return nil
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

func (d *Driver) getInstance() (*amz.EC2Instance, error) {
	instance, err := d.getClient().GetInstance(d.InstanceId)
	if err != nil {
		return nil, err
	}

	return &instance, nil
}

func (d *Driver) waitForInstance() error {
	for {
		st, err := d.GetState()
		if err != nil {
			return err
		}
		if st == state.Running {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if err := d.updateDriver(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) createKeyPair() error {

	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
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
		for {
			_, err := d.getClient().GetSecurityGroupById(group.GroupId)
			if err == nil {
				break
			}
			log.Debug(err)
			time.Sleep(1 * time.Second)
		}
	}

	d.SecurityGroupId = securityGroup.GroupId

	log.Debugf("configuring security group authorization for %s", ipRange)

	perms := configureSecurityGroupPermissions(securityGroup)

	if len(perms) != 0 {
		log.Debugf("authorizing group %s with permissions: %v", securityGroup.GroupName, perms)
		if err := d.getClient().AuthorizeSecurityGroup(d.SecurityGroupId, perms); err != nil {
			return err
		}

	}

	return nil
}

func configureSecurityGroupPermissions(group *amz.SecurityGroup) []amz.IpPermission {
	hasSshPort := false
	hasDockerPort := false
	for _, p := range group.IpPermissions {
		switch p.FromPort {
		case 22:
			hasSshPort = true
		case dockerPort:
			hasDockerPort = true
		}
	}

	perms := []amz.IpPermission{}

	if !hasSshPort {
		perm := amz.IpPermission{
			IpProtocol: "tcp",
			FromPort:   22,
			ToPort:     22,
			IpRange:    ipRange,
		}

		perms = append(perms, perm)
	}

	if !hasDockerPort {
		perm := amz.IpPermission{
			IpProtocol: "tcp",
			FromPort:   dockerPort,
			ToPort:     dockerPort,
			IpRange:    ipRange,
		}

		perms = append(perms, perm)
	}

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

package amazonec2

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/drivers/amazonec2/amz"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

const (
	driverName          = "amazonec2"
	defaultRegion       = "us-east-1"
	defaultAMI          = "ami-a00461c8"
	defaultInstanceType = "t2.micro"
	defaultRootSize     = 16
	ipRange             = "0.0.0.0/0"
)

type Driver struct {
	Id              string
	AccessKey       string
	SecretKey       string
	Region          string
	AMI             string
	SSHKeyID        int
	KeyName         string
	InstanceId      string
	InstanceType    string
	IPAddress       string
	MachineName     string
	SecurityGroupId string
	ReservationId   string
	RootSize        int64
	VpcId           string
	SubnetId        string
	Zone            string
	storePath       string
	keyPath         string
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
			Name:   "amazonec2-ami",
			Usage:  "AWS machine image",
			Value:  defaultAMI,
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

func NewDriver(machineName string, storePath string) (drivers.Driver, error) {
	id := generateId()
	return &Driver{Id: id, MachineName: machineName, storePath: storePath}, nil
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.AccessKey = flags.String("amazonec2-access-key")
	d.SecretKey = flags.String("amazonec2-secret-key")
	d.AMI = flags.String("amazonec2-ami")
	d.Region = flags.String("amazonec2-region")
	d.InstanceType = flags.String("amazonec2-instance-type")
	d.VpcId = flags.String("amazonec2-vpc-id")
	d.SubnetId = flags.String("amazonec2-subnet-id")
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

func (d *Driver) Create() error {
	log.Infof("Launching instance...")

	if err := d.createKeyPair(); err != nil {
		fmt.Errorf("unable to create key pair: %s", err)
	}

	group, err := d.createSecurityGroup()
	if err != nil {
		log.Fatalf("Please make sure you don't have a security group named: %s", d.MachineName)
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
	instance, err := d.getClient().RunInstance(d.AMI, d.InstanceType, d.Zone, 1, 1, group.GroupId, d.KeyName, subnetId, bdm)

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

	cmd, err = d.GetSSHCommand("if [ ! -e /usr/bin/docker ]; then curl get.docker.io | sudo sh -; fi")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = d.GetSSHCommand("sudo stop docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("HACK: Downloading version of Docker with identity auth...")

	cmd, err = d.GetSSHCommand("sudo curl -sS -o /usr/bin/docker https://ehazlett.s3.amazonaws.com/public/docker/linux/docker-1.4.1-136b351e-identity")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Debugf("Updating /etc/default/docker to use identity auth...")

	cmd, err = d.GetSSHCommand("echo 'export DOCKER_OPTS=\"--auth=identity --host=tcp://0.0.0.0:2376 --host=unix:///var/run/docker.sock --auth-authorized-dir=/root/.docker/authorized-keys.d\"' | sudo tee -a /etc/default/docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	// HACK: create dir for ubuntu user to access
	log.Debugf("Adding key to authorized-keys.d...")

	cmd, err = d.GetSSHCommand("sudo mkdir -p /root/.docker && sudo chown -R ubuntu /root/.docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	f, err := os.Open(filepath.Join(os.Getenv("HOME"), ".docker/public-key.json"))
	if err != nil {
		return err
	}
	defer f.Close()

	cmdString := fmt.Sprintf("sudo mkdir -p %q && sudo tee -a %q", "/root/.docker/authorized-keys.d", "/root/.docker/authorized-keys.d/docker-host.json")
	cmd, err = d.GetSSHCommand(cmdString)
	if err != nil {
		return err
	}
	cmd.Stdin = f
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd, err = d.GetSSHCommand("sudo start docker")
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
	return fmt.Sprintf("tcp://%s:2376", ip), nil
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
	// wait until terminated so we can remove security group
	for {
		st, err := d.GetState()
		if err != nil {
			break
		}
		if st == state.None {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if err := d.deleteSecurityGroup(); err != nil {
		return fmt.Errorf("unable to remove security group: %s", err)
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

func (d *Driver) Upgrade() error {
	sshCmd, err := d.GetSSHCommand("apt-get update && apt-get install -y lxc-docker")
	if err != nil {
		return err
	}
	sshCmd.Stdin = os.Stdin
	sshCmd.Stdout = os.Stdout
	sshCmd.Stderr = os.Stderr
	return sshCmd.Run()
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return ssh.GetSSHCommand(d.IPAddress, 22, "ubuntu", d.sshKeyPath(), args...), nil
}

func (d *Driver) getClient() *amz.EC2 {
	auth := amz.GetAuth(d.AccessKey, d.SecretKey)
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
	d.InstanceId = inst.InstanceId
	d.IPAddress = inst.IpAddress
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

func (d *Driver) createSecurityGroup() (*amz.SecurityGroup, error) {
	log.Debugf("creating security group in %s", d.VpcId)

	grpName := d.MachineName
	group, err := d.getClient().CreateSecurityGroup(grpName, "Docker Machine", d.VpcId)
	if err != nil {
		return nil, err
	}

	d.SecurityGroupId = group.GroupId

	perms := []amz.IpPermission{
		{
			Protocol: "tcp",
			FromPort: 22,
			ToPort:   22,
			IpRange:  ipRange,
		},
		{
			Protocol: "tcp",
			FromPort: 2376,
			ToPort:   2376,
			IpRange:  ipRange,
		},
	}

	log.Debugf("authorizing %s", ipRange)

	if err := d.getClient().AuthorizeSecurityGroup(d.SecurityGroupId, perms); err != nil {
		return nil, err
	}

	return group, nil
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

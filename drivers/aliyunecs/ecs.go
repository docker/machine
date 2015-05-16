package aliyunecs

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/denverdino/aliyungo/ecs"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/provider"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
	"io"
	"io/ioutil"

	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	//"time"
	//"os"
)

const (
	driverName               = "aliyunecs"
	defaultRegion            = "cn-hangzhou"
	defaultInstanceType      = "ecs.t1.small"
	defaultRootSize          = 20
	ipRange                  = "0.0.0.0/0"
	machineSecurityGroupName = "docker-machine"
)

var (
	dockerPort = 2376
	swarmPort  = 3376
)

type Driver struct {
	Id                      string
	AccessKey               string
	SecretKey               string
	Region                  ecs.Region
	ImageID                 string
	SSHKeyID                int
	SSHUser                 string
	SSHPassword             string
	SSHPort                 int
	PublicKey               []byte
	InstanceId              string
	InstanceType            string
	IPAddress               string
	PrivateIPAddress        string
	MachineName             string
	SecurityGroupId         string
	SecurityGroupName       string
	ReservationId           string
	VpcId                   string
	Zone                    string
	CaCertPath              string
	PrivateKeyPath          string
	SwarmMaster             bool
	SwarmHost               string
	SwarmDiscovery          string
	storePath               string
	keyPath                 string
	PrivateIPOnly           bool
	InternetMaxBandwidthOut int
	client                  *ecs.Client
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
			Name:   "aliyunecs-access-key",
			Usage:  "ECS Access Key",
			Value:  "",
			EnvVar: "ECS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "aliyunecs-secret-key",
			Usage:  "ECS Secret Key",
			Value:  "",
			EnvVar: "ECS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "aliyunecs-image-id",
			Usage:  "ECS machine image",
			EnvVar: "ECS_IMAGE_ID",
		},
		cli.StringFlag{
			Name:   "aliyunecs-region",
			Usage:  "ECS region, default cn-hangzhou",
			Value:  defaultRegion,
			EnvVar: "ECS_REGION",
		},
		cli.StringFlag{
			Name:   "aliyunecs-vpc-id",
			Usage:  "ECS VPC id",
			Value:  "",
			EnvVar: "ECS_VPC_ID",
		},
		cli.StringFlag{
			Name:   "aliyunecs-zone",
			Usage:  "ECS zone for instance",
			Value:  "",
			EnvVar: "ECS_ZONE",
		},
		cli.StringFlag{
			Name:   "aliyunecs-security-group",
			Usage:  "ECS VPC security group",
			Value:  "docker-machine",
			EnvVar: "ECS_SECURITY_GROUP",
		},
		cli.StringFlag{
			Name:   "aliyunecs-instance-type",
			Usage:  "ECS instance type",
			Value:  defaultInstanceType,
			EnvVar: "ECS_INSTANCE_TYPE",
		},
		//		cli.StringFlag{
		//			Name:   "aliyunecs-ssh-user",
		//			Usage:  "set the name of the ssh user",
		//			Value:  "root",
		//			EnvVar: "ECS_SSH_USER",
		//		},
		cli.StringFlag{
			Name:   "aliyunecs-ssh-password",
			Usage:  "set the password of the ssh user",
			EnvVar: "ECS_SSH_PASSWORD",
		},
		cli.BoolFlag{
			Name:  "aliyunecs-private-address-only",
			Usage: "Only use a private IP address",
		},
		cli.IntFlag{
			Name:   "aliyunecs-internet-max-bandwidth",
			Usage:  "Maxium bandwidth for Internet access (in Mbps), default 1",
			Value:  1,
			EnvVar: "ECS_INTERNET_MAX_BANDWIDTH",
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

func (d *Driver) GetImageID(image string) string {

	if len(image) != 0 {
		return image
	}
	args := ecs.DescribeImagesArgs{
		RegionId:        d.Region,
		ImageOwnerAlias: ecs.ImageOwnerSystem,
	}

	// Scan registed images with prefix of ubuntu1404_64_20G_
	for {
		images, pagination, err := d.getClient().DescribeImages(&args)
		if err != nil {
			log.Errorf("Failed to describe images: %v", err)
			break
		} else {
			for _, image := range images {
				if strings.HasPrefix(image.ImageId, defaultUbuntuImagePrefix) {
					return image.ImageId
				}
			}
			nextPage := pagination.NextPage()
			if nextPage == nil {
				break
			}
			args.Pagination = *nextPage
		}
	}

	//Default use the config Ubuntu 14.04 64bits image

	image = defaultUbuntuImageID

	return image
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	region, err := validateECSRegion(flags.String("aliyunecs-region"))
	if err != nil {
		return err
	}
	d.AccessKey = flags.String("aliyunecs-access-key")
	d.SecretKey = flags.String("aliyunecs-secret-key")
	d.Region = region
	d.ImageID = flags.String("aliyunecs-image-id")
	d.InstanceType = flags.String("aliyunecs-instance-type")
	d.VpcId = flags.String("aliyunecs-vpc-id")
	d.SecurityGroupName = flags.String("aliyunecs-security-group")
	zone := flags.String("aliyunecs-zone")
	d.Zone = zone[:]
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = "root" //TODO support non-root
	d.SSHPassword = flags.String("aliyunecs-ssh-password")
	d.SSHPort = 22
	d.PrivateIPOnly = flags.Bool("aliyunecs-private-address-only")
	d.InternetMaxBandwidthOut = flags.Int("aliyunecs-internet-max-bandwidth")

	//TODO support PayByTraffic
	if d.InternetMaxBandwidthOut < 0 || d.InternetMaxBandwidthOut > 100 {
		return fmt.Errorf("aliyunecs driver invalid --aliyunecs-internet-max-bandwidth the value should be in 1 ~ 100")
	}

	if d.InternetMaxBandwidthOut == 0 {
		d.InternetMaxBandwidthOut = 1
	}

	if d.AccessKey == "" {
		return fmt.Errorf("aliyunecs driver requires the --aliyunecs-access-key option")
	}

	if d.SecretKey == "" {
		return fmt.Errorf("aliyunecs driver requires the --aliyunecs-secret-key option")
	}

	// VpcId is optional

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
	// TODO
	return nil
}

func (d *Driver) PreCreateCheck() error {
	return d.checkPrereqs()
}

func (d *Driver) Create() error {
	if err := d.checkPrereqs(); err != nil {
		return err
	}
	log.Infof("Creating key pair for instances...")

	if err := d.createKeyPair(); err != nil {
		return fmt.Errorf("unable to create key pair: %s", err)
	}

	log.Infof("Configuring security groups...")
	if err := d.configureSecurityGroup(d.SecurityGroupName); err != nil {
		return err
	}

	// TODO Support data disk
	//	bdm := &amz.BlockDeviceMapping{
	//		DeviceName:          "/dev/sda1",
	//		VolumeSize:          d.RootSize,
	//		DeleteOnTermination: true,
	//		VolumeType:          "gp2",
	//	}

	if d.SSHPassword == "" {
		d.SSHPassword = randomPassword()
		log.Info("Launching instance with generated password, please update password in console or log in with ssh key.")
	}

	imageID := d.GetImageID(d.ImageID)
	log.Infof("Launching instance with image %s ...", imageID)

	args := ecs.CreateInstanceArgs{
		RegionId:                d.Region,
		ImageId:                 imageID,
		InstanceType:            d.InstanceType,
		SecurityGroupId:         d.SecurityGroupId,
		InternetMaxBandwidthOut: d.InternetMaxBandwidthOut,
		Password:                d.SSHPassword,
	}

	// Create instance
	instanceId, err := d.getClient().CreateInstance(&args)

	if err != nil {
		return fmt.Errorf("Error create instance: %s", err)
	}

	d.InstanceId = instanceId

	// Wait for creation successfully
	err = d.getClient().WaitForInstance(instanceId, ecs.Stopped, 60)

	if err != nil {
		return fmt.Errorf("Error wait instance to Stopped: %s", err)
	}

	// Assign public IP if not private IP only
	if !d.PrivateIPOnly {
		_, err := d.getClient().AllocatePublicIpAddress(instanceId)
		if err != nil {
			return fmt.Errorf("Error allocate public IP address for instance %s: %v", instanceId, err)
		}
	}

	// Start instance
	err = d.getClient().StartInstance(instanceId)
	if err != nil {
		return fmt.Errorf("Error start instance %s: %v", instanceId, err)
	}

	// Wait for running
	err = d.getClient().WaitForInstance(instanceId, ecs.Running, 300)

	if err != nil {
		return fmt.Errorf("Error wait instance to Running: %s", err)
	}

	instance, err := d.getInstance()

	if err != nil {
		return err
	}

	if len(instance.InnerIpAddress.IpAddress) > 0 {
		d.PrivateIPAddress = instance.InnerIpAddress.IpAddress[0]
	}

	d.IPAddress = d.getIP(instance)

	d.uploadKeyPair()

	log.Debugf("created instance ID %s, IP address %s, Private IP address %s",
		d.InstanceId,
		d.IPAddress,
		d.PrivateIPAddress,
	)

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

	return d.getIP(inst), nil
}

func (d *Driver) getIP(inst *ecs.InstanceAttributesType) string {

	if d.PrivateIPOnly {
		if inst.InnerIpAddress.IpAddress != nil && len(inst.InnerIpAddress.IpAddress) > 0 {
			return inst.InnerIpAddress.IpAddress[0]
		}
	}
	if inst.PublicIpAddress.IpAddress != nil && len(inst.PublicIpAddress.IpAddress) > 0 {
		return inst.PublicIpAddress.IpAddress[0]
	}
	return ""
}

func (d *Driver) GetState() (state.State, error) {
	inst, err := d.getInstance()
	if err != nil {
		return state.Error, err
	}
	switch ecs.InstanceStatus(inst.Status) {
	case ecs.Starting:
		return state.Starting, nil
	case ecs.Running:
		return state.Running, nil
	case ecs.Stopping:
		return state.Stopping, nil
	case ecs.Stopped:
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
		d.SSHUser = "root"
	}

	return d.SSHUser
}

func (d *Driver) Start() error {
	if err := d.getClient().StartInstance(d.InstanceId); err != nil {
		return err
	}

	// Wait for running
	err := d.getClient().WaitForInstance(d.InstanceId, ecs.Running, 300)

	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Stop() error {
	if err := d.getClient().StopInstance(d.InstanceId, false); err != nil {
		return err
	}

	// Wait for stopped
	err := d.getClient().WaitForInstance(d.InstanceId, ecs.Stopped, 300)

	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Remove() error {
	if d.InstanceId == "" {
		return fmt.Errorf("unknown instance")
	}

	s, err := d.GetState()
	if err == nil && s == state.Running {
		if err := d.Stop(); err != nil {
			log.Errorf("unable to stop instance: %s", err)
		}
	}

	log.Debugf("terminating instance: %s", d.InstanceId)
	if err := d.getClient().DeleteInstance(d.InstanceId); err != nil {
		return fmt.Errorf("unable to terminate instance: %s", err)
	}
	return nil
}

func (d *Driver) Restart() error {
	if err := d.getClient().RebootInstance(d.InstanceId, false); err != nil {
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

func (d *Driver) getClient() *ecs.Client {
	if d.client == nil {
		client := ecs.NewClient(d.AccessKey, d.SecretKey)
		client.SetDebug(false)
		d.client = client
	}
	return d.client
}

func (d *Driver) GetSSHKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) getInstance() (*ecs.InstanceAttributesType, error) {
	return d.getClient().DescribeInstanceAttribute(d.InstanceId)
}

func (d *Driver) createKeyPair() error {

	log.Debug("SSH key path: ", d.GetSSHKeyPath())

	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	publicKey, err := ioutil.ReadFile(d.GetSSHKeyPath() + ".pub")
	if err != nil {
		return err
	}

	d.PublicKey = publicKey
	return nil
}

func (d *Driver) isSwarmMaster() bool {
	return d.SwarmMaster
}

func (d *Driver) getSecurityGroup(id string) (sg *ecs.DescribeSecurityGroupAttributeResponse, err error) {
	args := ecs.DescribeSecurityGroupAttributeArgs{
		SecurityGroupId: id,
		RegionId:        d.Region,
	}
	return d.getClient().DescribeSecurityGroupAttribute(&args)
}

func (d *Driver) securityGroupAvailableFunc(id string) func() bool {
	return func() bool {
		_, err := d.getSecurityGroup(id)
		if err == nil {
			return true
		}
		log.Debug(err)
		return false
	}
}

func (d *Driver) configureSecurityGroup(groupName string) error {
	log.Debugf("configuring security group in %s", d.VpcId)

	var securityGroup *ecs.DescribeSecurityGroupAttributeResponse

	args := ecs.DescribeSecurityGroupsArgs{
		RegionId: d.Region,
		VpcId:    d.VpcId,
	}

	//TODO handle pagination
	groups, _, err := d.getClient().DescribeSecurityGroups(&args)
	if err != nil {
		return err
	}

	log.Debugf("DescribeSecurityGroups: %++v\n", groups)

	for _, grp := range groups {
		if grp.SecurityGroupName == groupName {
			log.Debugf("found existing security group (%s) in %s", groupName, d.VpcId)
			securityGroup, _ = d.getSecurityGroup(grp.SecurityGroupId)
			break
		}
	}

	// if not found, create
	if securityGroup == nil {
		log.Debugf("creating security group (%s) in %s", groupName, d.VpcId)
		creationArgs := ecs.CreateSecurityGroupArgs{
			RegionId:          d.Region,
			SecurityGroupName: groupName,
			Description:       "Docker Machine",
			VpcId:             d.VpcId,
			ClientToken:       d.getClient().GenerateClientToken(),
		}

		groupId, err := d.getClient().CreateSecurityGroup(&creationArgs)
		if err != nil {
			return err
		}

		// wait until created (dat eventual consistency)
		log.Debugf("waiting for group (%s) to become available", groupId)
		if err := utils.WaitFor(d.securityGroupAvailableFunc(groupId)); err != nil {
			return err
		}
		securityGroup, err = d.getSecurityGroup(groupId)
		if err != nil {
			return err
		}
	}

	d.SecurityGroupId = securityGroup.SecurityGroupId

	perms := d.configureSecurityGroupPermissions(securityGroup)

	for _, permission := range perms {
		log.Debugf("authorizing group %s with permission: %v", securityGroup.SecurityGroupName, permission)
		args := permission.createAuthorizeSecurityGroupArgs(d.Region, d.SecurityGroupId)
		if err := d.getClient().AuthorizeSecurityGroup(args); err != nil {
			return err
		}

	}

	return nil
}

type IpPermission struct {
	IpProtocol ecs.IpProtocol
	FromPort   int
	ToPort     int
	IpRange    string
}

func (p *IpPermission) createAuthorizeSecurityGroupArgs(regionId ecs.Region, securityGroupId string) *ecs.AuthorizeSecurityGroupArgs {
	args := ecs.AuthorizeSecurityGroupArgs{
		RegionId:        regionId,
		SecurityGroupId: securityGroupId,
		IpProtocol:      p.IpProtocol,
		SourceCidrIp:    p.IpRange,
		PortRange:       fmt.Sprintf("%d/%d", p.FromPort, p.ToPort),
	}
	return &args
}

func (d *Driver) configureSecurityGroupPermissions(group *ecs.DescribeSecurityGroupAttributeResponse) []IpPermission {
	hasSshPort := false
	hasDockerPort := false
	hasSwarmPort := false
	for _, p := range group.Permissions.Permission {
		portRange := strings.Split(p.PortRange, "/")
		log.Debug("portRange", portRange)
		fromPort, _ := strconv.Atoi(portRange[0])
		switch fromPort {
		case 22:
			hasSshPort = true
		case dockerPort:
			hasDockerPort = true
		case swarmPort:
			hasSwarmPort = true
		}
	}

	perms := []IpPermission{}

	if !hasSshPort {
		perms = append(perms, IpPermission{
			IpProtocol: ecs.IpProtocolTCP,
			FromPort:   22,
			ToPort:     22,
			IpRange:    ipRange,
		})
	}

	if !hasDockerPort {
		perms = append(perms, IpPermission{
			IpProtocol: "tcp",
			FromPort:   dockerPort,
			ToPort:     dockerPort,
			IpRange:    ipRange,
		})
	}

	if !hasSwarmPort && d.SwarmMaster {
		perms = append(perms, IpPermission{
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

	if err := d.getClient().DeleteSecurityGroup(d.Region, d.SecurityGroupId); err != nil {
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

func (d *Driver) uploadKeyPair() error {

	ipAddr := d.IPAddress
	port, _ := d.GetSSHPort()
	tcpAddr := fmt.Sprintf("%s:%d", ipAddr, port)

	ssh.WaitForTCP(tcpAddr)

	auth := ssh.Auth{
		Passwords: []string{d.SSHPassword},
	}

	sshClient, err := ssh.NewClient(d.GetSSHUsername(), ipAddr, port, &auth)

	if err != nil {
		return err
	}

	command := fmt.Sprintf("mkdir -p ~/.ssh; echo '%s' > ~/.ssh/authorized_keys", string(d.PublicKey))

	log.Debugf("Upload the public key with command: %s", command)

	output, err := sshClient.Run(command)

	log.Debugf(fmt.Sprintf("Upload command err, output: %v: %s", err, output))

	if err != nil {
		return err
	}

	log.Debugf("Upload the public key with command: %s", command)

	// Fix the routing rule
	output, err = sshClient.Run("route del -net 172.16.0.0/12")
	log.Debugf(fmt.Sprintf("Delete route command err, output: %v: %s", err, output))

	output, err = sshClient.Run("sed -i -r 's/^(up route add \\-net 172\\..*)$/#\\1/' /etc/network/interfaces")
	log.Debugf(fmt.Sprintf("Fix route in /etc/network/interfaces command err, output: %v: %s", err, output))

	return nil
}

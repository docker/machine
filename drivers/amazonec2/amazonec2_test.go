package amazonec2

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/docker/machine/commands/commandstest"
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/stretchr/testify/assert"
)

const (
	testSSHPort           = 22
	testDockerPort        = 2376
	testStoreDir          = ".store-test"
	machineTestName       = "test-host"
	machineTestDriverName = "none"
	machineTestStorePath  = "/test/path"
	machineTestCaCert     = "test-cert"
	machineTestPrivateKey = "test-key"
)

var (
	securityGroup = &ec2.SecurityGroup{
		GroupName: aws.String("test-group"),
		GroupId:   aws.String("12345"),
		VpcId:     aws.String("12345"),
	}
)

func cleanup() error {
	return os.RemoveAll(testStoreDir)
}

func getTestStorePath() (string, error) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		return "", err
	}
	mcndirs.BaseDir = tmpDir
	return tmpDir, nil
}

func getDefaultTestDriverFlags() *commandstest.FakeFlagger {
	return &commandstest.FakeFlagger{
		Data: map[string]interface{}{
			"name":                            "test",
			"url":                             "unix:///var/run/docker.sock",
			"swarm":                           false,
			"swarm-host":                      "",
			"swarm-master":                    false,
			"swarm-discovery":                 "",
			"amazonec2-ami":                   "ami-12345",
			"amazonec2-access-key":            "abcdefg",
			"amazonec2-secret-key":            "12345",
			"amazonec2-session-token":         "",
			"amazonec2-instance-type":         "t1.micro",
			"amazonec2-vpc-id":                "vpc-12345",
			"amazonec2-subnet-id":             "subnet-12345",
			"amazonec2-security-group":        "docker-machine-test",
			"amazonec2-region":                "us-east-1",
			"amazonec2-zone":                  "e",
			"amazonec2-root-size":             10,
			"amazonec2-iam-instance-profile":  "",
			"amazonec2-ssh-user":              "ubuntu",
			"amazonec2-request-spot-instance": false,
			"amazonec2-spot-price":            "",
			"amazonec2-private-address-only":  false,
			"amazonec2-use-private-address":   false,
			"amazonec2-monitoring":            false,
		},
	}
}

func getTestDriver() (*Driver, error) {
	storePath, err := getTestStorePath()
	if err != nil {
		return nil, err
	}
	defer cleanup()

	d := NewDriver(machineTestName, storePath)
	d.SetConfigFromFlags(getDefaultTestDriverFlags())
	drv := d.(*Driver)
	return drv, nil
}

func TestConfigureSecurityGroupPermissionsEmpty(t *testing.T) {
	d, err := getTestDriver()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	group := securityGroup
	perms := d.configureSecurityGroupPermissions(group)
	if len(perms) != 2 {
		t.Fatalf("expected 2 permissions; received %d", len(perms))
	}
}

func TestConfigureSecurityGroupPermissionsSshOnly(t *testing.T) {
	d, err := getTestDriver()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	group := securityGroup

	group.IpPermissions = []*ec2.IpPermission{
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(int64(testSSHPort)),
			ToPort:     aws.Int64(int64(testSSHPort)),
		},
	}

	perms := d.configureSecurityGroupPermissions(group)
	if len(perms) != 1 {
		t.Fatalf("expected 1 permission; received %d", len(perms))
	}

	receivedPort := *perms[0].FromPort
	if receivedPort != testDockerPort {
		t.Fatalf("expected permission on port %d; received port %d", testDockerPort, receivedPort)
	}
}

func TestConfigureSecurityGroupPermissionsDockerOnly(t *testing.T) {
	d, err := getTestDriver()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	group := securityGroup

	group.IpPermissions = []*ec2.IpPermission{
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64((testDockerPort)),
			ToPort:     aws.Int64((testDockerPort)),
		},
	}

	perms := d.configureSecurityGroupPermissions(group)
	if len(perms) != 1 {
		t.Fatalf("expected 1 permission; received %d", len(perms))
	}

	receivedPort := *perms[0].FromPort
	if receivedPort != testSSHPort {
		t.Fatalf("expected permission on port %d; received port %d", testSSHPort, receivedPort)
	}
}

func TestConfigureSecurityGroupPermissionsDockerAndSsh(t *testing.T) {
	d, err := getTestDriver()
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	group := securityGroup

	group.IpPermissions = []*ec2.IpPermission{
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(testSSHPort),
			ToPort:     aws.Int64(testSSHPort),
		},
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(testDockerPort),
			ToPort:     aws.Int64(testDockerPort),
		},
	}

	perms := d.configureSecurityGroupPermissions(group)
	if len(perms) != 0 {
		t.Fatalf("expected 0 permissions; received %d", len(perms))
	}
}

func TestAwsRegionList(t *testing.T) {
}

func TestValidateAwsRegionValid(t *testing.T) {
	regions := []string{"eu-west-1", "eu-central-1"}

	for _, v := range regions {
		r, err := validateAwsRegion(v)
		if err != nil {
			t.Fatal(err)
		}

		if v != r {
			t.Fatal("Wrong region returned")
		}
	}
}

func TestValidateAwsRegionInvalid(t *testing.T) {
	regions := []string{"eu-west-2", "eu-central-2"}

	for _, v := range regions {
		r, err := validateAwsRegion(v)
		if err == nil {
			t.Fatal("No error returned")
		}

		if v == r {
			t.Fatal("Wrong region returned")
		}
	}
}

func TestSetConfigFromFlags(t *testing.T) {
	driver, err := getTestDriver()
	if err != nil {
		t.Fatal(err)
	}

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"amazonec2-region": "us-west-2",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}

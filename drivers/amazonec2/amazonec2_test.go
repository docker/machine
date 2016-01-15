package amazonec2

import (
	"testing"

	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/docker/machine/commands/commandstest"
	"github.com/stretchr/testify/assert"
)

const (
	testSSHPort    = 22
	testDockerPort = 2376
	testSwarmPort  = 3376
)

var (
	securityGroup = &ec2.SecurityGroup{
		GroupName: aws.String("test-group"),
		GroupId:   aws.String("12345"),
		VpcId:     aws.String("12345"),
	}
)

func TestConfigureSecurityGroupPermissionsEmpty(t *testing.T) {
	driver := NewTestDriver()

	perms := driver.configureSecurityGroupPermissions(securityGroup)

	assert.Len(t, perms, 2)
}

func TestConfigureSecurityGroupPermissionsSshOnly(t *testing.T) {
	driver := NewTestDriver()
	group := securityGroup
	group.IpPermissions = []*ec2.IpPermission{
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64(int64(testSSHPort)),
			ToPort:     aws.Int64(int64(testSSHPort)),
		},
	}

	perms := driver.configureSecurityGroupPermissions(group)

	assert.Len(t, perms, 1)
	assert.Equal(t, testDockerPort, *perms[0].FromPort)
}

func TestConfigureSecurityGroupPermissionsDockerOnly(t *testing.T) {
	driver := NewTestDriver()
	group := securityGroup
	group.IpPermissions = []*ec2.IpPermission{
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int64((testDockerPort)),
			ToPort:     aws.Int64((testDockerPort)),
		},
	}

	perms := driver.configureSecurityGroupPermissions(group)

	assert.Len(t, perms, 1)
	assert.Equal(t, testSSHPort, *perms[0].FromPort)
}

func TestConfigureSecurityGroupPermissionsDockerAndSsh(t *testing.T) {
	driver := NewTestDriver()
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

	perms := driver.configureSecurityGroupPermissions(group)

	assert.Empty(t, perms)
}

func TestConfigureSecurityGroupPermissionsWithSwarm(t *testing.T) {
	driver := NewTestDriver()
	driver.SwarmMaster = true
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

	perms := driver.configureSecurityGroupPermissions(group)

	assert.Len(t, perms, 1)
	assert.Equal(t, testSwarmPort, *perms[0].FromPort)
}

func TestValidateAwsRegionValid(t *testing.T) {
	regions := []string{"eu-west-1", "eu-central-1"}

	for _, region := range regions {
		validatedRegion, err := validateAwsRegion(region)

		assert.NoError(t, err)
		assert.Equal(t, region, validatedRegion)
	}
}

func TestValidateAwsRegionInvalid(t *testing.T) {
	regions := []string{"eu-west-2", "eu-central-2"}

	for _, region := range regions {
		_, err := validateAwsRegion(region)

		assert.EqualError(t, err, "Invalid region specified")
	}
}

func TestFindDefaultVPC(t *testing.T) {
	driver := NewDriver("machineFoo", "path")
	driver.clientFactory = func() Ec2Client { return &fakeEC2WithLogin{} }

	vpc, err := driver.getDefaultVPCId()

	assert.Equal(t, "vpc-9999", vpc)
	assert.NoError(t, err)
}

func TestDefaultVPCIsMissing(t *testing.T) {
	driver := NewDriver("machineFoo", "path")
	driver.clientFactory = func() Ec2Client {
		return &fakeEC2WithDescribe{
			output: &ec2.DescribeAccountAttributesOutput{
				AccountAttributes: []*ec2.AccountAttribute{},
			},
		}
	}

	vpc, err := driver.getDefaultVPCId()

	assert.EqualError(t, err, "No default-vpc attribute")
	assert.Empty(t, vpc)
}

func TestDescribeAccountAttributeFails(t *testing.T) {
	driver := NewDriver("machineFoo", "path")
	driver.clientFactory = func() Ec2Client {
		return &fakeEC2WithDescribe{
			err: errors.New("Not Found"),
		}
	}

	vpc, err := driver.getDefaultVPCId()

	assert.EqualError(t, err, "Not Found")
	assert.Empty(t, vpc)
}

func TestAccessKeyIsMandatory(t *testing.T) {
	driver := NewTestDriver()
	driver.awsCredentials = &cliCredentials{}
	options := &commandstest.FakeFlagger{
		Data: map[string]interface{}{
			"name":             "test",
			"amazonec2-region": "us-east-1",
			"amazonec2-zone":   "e",
		},
	}

	err := driver.SetConfigFromFlags(options)

	assert.Equal(t, err, errorMissingAccessKeyOption)
}

func TestAccessKeyIsMandatoryEvenIfSecretKeyIsPassed(t *testing.T) {
	driver := NewTestDriver()
	driver.awsCredentials = &cliCredentials{}
	options := &commandstest.FakeFlagger{
		Data: map[string]interface{}{
			"name":                 "test",
			"amazonec2-secret-key": "123",
			"amazonec2-region":     "us-east-1",
			"amazonec2-zone":       "e",
		},
	}

	err := driver.SetConfigFromFlags(options)

	assert.Equal(t, err, errorMissingAccessKeyOption)
}

func TestSecretKeyIsMandatory(t *testing.T) {
	driver := NewTestDriver()
	driver.awsCredentials = &cliCredentials{}
	options := &commandstest.FakeFlagger{
		Data: map[string]interface{}{
			"name":                 "test",
			"amazonec2-access-key": "foobar",
			"amazonec2-region":     "us-east-1",
			"amazonec2-zone":       "e",
		},
	}

	err := driver.SetConfigFromFlags(options)

	assert.Equal(t, err, errorMissingSecretKeyOption)
}

func TestLoadingFromCredentialsWorked(t *testing.T) {
	driver := NewCustomTestDriver(&fakeEC2WithLogin{})
	driver.awsCredentials = &fileCredentials{}
	options := &commandstest.FakeFlagger{
		Data: map[string]interface{}{
			"name":             "test",
			"amazonec2-region": "us-east-1",
			"amazonec2-zone":   "e",
		},
	}

	err := driver.SetConfigFromFlags(options)

	assert.NoError(t, err)
	assert.Equal(t, "access", driver.AccessKey)
	assert.Equal(t, "secret", driver.SecretKey)
	assert.Equal(t, "token", driver.SessionToken)
}

func TestPassingBothCLIArgWorked(t *testing.T) {
	driver := NewCustomTestDriver(&fakeEC2WithLogin{})
	driver.awsCredentials = &cliCredentials{}
	options := &commandstest.FakeFlagger{
		Data: map[string]interface{}{
			"name":                 "test",
			"amazonec2-access-key": "foobar",
			"amazonec2-secret-key": "123",
			"amazonec2-region":     "us-east-1",
			"amazonec2-zone":       "e",
		},
	}

	err := driver.SetConfigFromFlags(options)

	assert.NoError(t, err)
	assert.Equal(t, "foobar", driver.AccessKey)
	assert.Equal(t, "123", driver.SecretKey)
}

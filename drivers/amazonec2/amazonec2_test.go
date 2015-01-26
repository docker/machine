package amazonec2

import (
	"testing"

	"github.com/docker/machine/drivers/amazonec2/amz"
)

var (
	securityGroup = amz.SecurityGroup{
		GroupName: "test-group",
		GroupId:   "12345",
		VpcId:     "12345",
	}
)

const (
	testSshPort    = 22
	testDockerPort = 2376
)

func TestConfigureSecurityGroupPermissionsEmpty(t *testing.T) {
	group := securityGroup
	perms := configureSecurityGroupPermissions(&group)
	if len(perms) != 2 {
		t.Fatalf("expected 2 permissions; received %d", len(perms))
	}
}

func TestConfigureSecurityGroupPermissionsSshOnly(t *testing.T) {
	group := securityGroup

	group.IpPermissions = []amz.IpPermission{
		{
			IpProtocol: "tcp",
			FromPort:   testSshPort,
			ToPort:     testSshPort,
		},
	}

	perms := configureSecurityGroupPermissions(&group)
	if len(perms) != 1 {
		t.Fatalf("expected 1 permission; received %d", len(perms))
	}

	receivedPort := perms[0].FromPort
	if receivedPort != testDockerPort {
		t.Fatalf("expected permission on port %d; received port %d", testDockerPort, receivedPort)
	}
}

func TestConfigureSecurityGroupPermissionsDockerOnly(t *testing.T) {
	group := securityGroup

	group.IpPermissions = []amz.IpPermission{
		{
			IpProtocol: "tcp",
			FromPort:   testDockerPort,
			ToPort:     testDockerPort,
		},
	}

	perms := configureSecurityGroupPermissions(&group)
	if len(perms) != 1 {
		t.Fatalf("expected 1 permission; received %d", len(perms))
	}

	receivedPort := perms[0].FromPort
	if receivedPort != testSshPort {
		t.Fatalf("expected permission on port %d; received port %d", testSshPort, receivedPort)
	}
}

func TestConfigureSecurityGroupPermissionsDockerAndSsh(t *testing.T) {
	group := securityGroup

	group.IpPermissions = []amz.IpPermission{
		{
			IpProtocol: "tcp",
			FromPort:   testSshPort,
			ToPort:     testSshPort,
		},
		{
			IpProtocol: "tcp",
			FromPort:   testDockerPort,
			ToPort:     testDockerPort,
		},
	}

	perms := configureSecurityGroupPermissions(&group)
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

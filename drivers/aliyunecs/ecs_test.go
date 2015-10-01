package aliyunecs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/denverdino/aliyungo/ecs"
)

const (
	testStoreDir    = ".store-test"
	machineTestName = "test-host"
)

func TestRandomPassword(t *testing.T) {

	for i := 0; i < 10; i++ {
		t.Logf("Random Password: %s", randomPassword())
	}

}

type DriverOptionsMock struct {
	Data map[string]interface{}
}

func (d DriverOptionsMock) String(key string) string {
	v := d.Data[key]
	if v == nil {
		v = ""
	}
	return v.(string)
}

func (d DriverOptionsMock) StringSlice(key string) []string {
	return d.Data[key].([]string)
}

func (d DriverOptionsMock) Int(key string) int {
	v := d.Data[key]
	if v == nil {
		v = 0
	}
	return v.(int)
}

func (d DriverOptionsMock) Bool(key string) bool {
	v := d.Data[key]
	if v == nil {
		v = false
	}
	return v.(bool)
}

func cleanup() error {
	return os.RemoveAll(testStoreDir)
}

func getTestStorePath() (string, error) {
	tmpDir, err := ioutil.TempDir("", "machine-test-")
	if err != nil {
		return "", err
	}
	os.Setenv("MACHINE_STORAGE_PATH", tmpDir)
	return tmpDir, nil
}

func getDefaultTestDriverFlags() *DriverOptionsMock {
	return &DriverOptionsMock{
		Data: map[string]interface{}{
			"name":             "test",
			"url":              "unix:///var/run/docker.sock",
			"swarm":            false,
			"swarm-host":       "",
			"swarm-master":     false,
			"swarm-discovery":  "",
			"aliyunecs-region": "cn-hangzhou",

			"aliyunecs-image-id":          "img-12345",
			"aliyunecs-access-key-id":     "abcdefg",
			"aliyunecs-access-key-secret": "12345",
		},
	}
}

func TestSetConfigFromFlags(t *testing.T) {
	flags := &DriverOptionsMock{
		Data: map[string]interface{}{
			"name":                    "test",
			"url":                     "unix:///var/run/docker.sock",
			"swarm":                   false,
			"swarm-host":              "",
			"swarm-master":            false,
			"swarm-discovery":         "",
			"aliyunecs-access-key-id": "abcdefg",
		},
	}

	storePath, err := getTestStorePath()
	if err != nil {
		return
	}
	defer cleanup()

	d := NewDriver(machineTestName, storePath)
	err = d.SetConfigFromFlags(flags)
	if err != nil {
		t.Logf("Error: %v", err)
	} else {
		t.Fatalf("SetConfigFromFlags failed")
	}
	flags.Data["aliyunecs-region"] = "cn-hangzhou"
	err = d.SetConfigFromFlags(flags)
	if err != nil {
		t.Logf("Error: %v", err)
	} else {
		t.Fatalf("SetConfigFromFlags failed")
	}
	flags.Data["aliyunecs-access-key-secret"] = "12345"
	err = d.SetConfigFromFlags(flags)

	if err != nil {
		t.Fatalf("SetConfigFromFlags should have no error")
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

	securityGroup := ecs.DescribeSecurityGroupAttributeResponse{
		SecurityGroupName: "test-group",
		SecurityGroupId:   "12345",
		VpcId:             "12345",
	}
	perms := d.configureSecurityGroupPermissions(&securityGroup)
	t.Logf("Permissions: %v", perms)
	if len(perms) != 3 {
		t.Fatalf("expected 2 permissions; received %d", len(perms))
	}
}

func TestGetPrivateIP(t *testing.T) {
	d, err := getTestDriver()
	if err != nil {
		t.Fatal(err)
	}

	var instance ecs.InstanceAttributesType
	instance.InnerIpAddress.IpAddress = []string{"192.168.1.100"}

	if d.getPrivateIP(&instance) != "192.168.1.100" {
		t.Error("getPrivateIP failed")
	}

	var instance2 ecs.InstanceAttributesType
	instance2.VpcAttributes.PrivateIpAddress.IpAddress = []string{"172.168.1.100"}

	if d.getPrivateIP(&instance2) != "172.168.1.100" {
		t.Error("getPrivateIP failed")
	}

}

func TestGetIP(t *testing.T) {
	d, err := getTestDriver()
	if err != nil {
		t.Fatal(err)
	}

	var instance ecs.InstanceAttributesType
	instance.InnerIpAddress.IpAddress = []string{"192.168.1.100"}
	instance.PublicIpAddress.IpAddress = []string{"42.120.158.67"}
	if d.getIP(&instance) != "42.120.158.67" {
		t.Error("getIP failed")
	}

	var instance2 ecs.InstanceAttributesType
	instance2.InnerIpAddress.IpAddress = []string{"192.168.1.100"}
	instance2.EipAddress.IpAddress = "42.120.158.67"
	if d.getIP(&instance2) != "42.120.158.67" {
		t.Error("getIP failed")
	}

	d.PrivateIPOnly = true
	if d.getIP(&instance) != "192.168.1.100" {
		t.Error("getIP failed")
	}
	if d.getIP(&instance2) != "192.168.1.100" {
		t.Error("getIP failed")
	}
}

func TestValidateECSRegionValid(t *testing.T) {

	validRegions := []string{"cn-beijing", "us-west-1"}

	for _, v := range validRegions {
		_, err := validateECSRegion(v)
		if err != nil {
			t.Fatalf("No error returned")
		}
	}

}

func TestValidateECSRegionInvalid(t *testing.T) {
	invalidRegions := []string{"", "cn-beijing-", "us-west-@"}

	for _, v := range invalidRegions {
		_, err := validateECSRegion(v)
		if err == nil {
			t.Fatalf("%s should be invalid region. ", v)
		}
	}

}

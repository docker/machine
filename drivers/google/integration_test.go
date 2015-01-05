package google

import (
	"flag"
	"io/ioutil"
	"os"
	"path"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/machine/state"
)

var (
	project   = flag.String("project", "", "Project")
	tokenPath = flag.String("token-path", "", "Token path")
)

var (
	driver *Driver
	c      *ComputeUtil
)

const (
	zone = "us-central1-a"
)

func init() {
	flag.Parse()

	if *project == "" {
		log.Error("You must specify a GCE project using the --project flag. All tests will be skipped.")
		return
	}

	if *tokenPath == "" {
		log.Error("You must specify a token using the --token-path flag. All tests will be skipped.")
		return
	}

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}

	// go test can't take args from stdin, so the path to an existing token must be passed as a flag.
	os.Link(*tokenPath, path.Join(tmpDir, "gce_token"))

	log.Fatal("hai")

	driver = &Driver{
		storePath:        tmpDir,
		MachineName:      "test-instance",
		Zone:             "us-central1-a",
		MachineType:      "n1-standard-1",
		UserName:         os.Getenv("USER"),
		Project:          *project,
		sshKeyPath:       path.Join(tmpDir, "id_rsa"),
		publicSSHKeyPath: path.Join(tmpDir, "id_rsa.pub"),
	}

	c, err = newComputeUtil(driver)
	if err != nil {
		log.Fatal(err)
	}
}

func cleanupDisk() {
	log.Println("Cleaning up disk.")
	d, err := c.service.Disks.Get(*project, zone, "test-instance-disk").Do()
	if d == nil {
		return
	}
	op, err := c.service.Disks.Delete(*project, zone, "test-instance-disk").Do()
	if err != nil {
		log.Printf("Error cleaning up disk: %v", err)
		return
	}
	err = c.waitForRegionalOp(op.Name)
	if err != nil {
		log.Printf("Error cleaning up disk: %v", err)
	}
}

func cleanupInstance() {
	log.Println("Cleaning up instance.")
	d, err := c.service.Instances.Get(*project, zone, "test-instance").Do()
	if d == nil {
		return
	}
	op, err := c.service.Instances.Delete(*project, zone, "test-instance").Do()
	if err != nil {
		log.Printf("Error cleaning up instance: %v", err)
		return
	}
	err = c.waitForRegionalOp(op.Name)
	if err != nil {
		log.Printf("Error cleaning up instance: %v", err)
	}
}

func cleanup() {
	cleanupInstance()
	cleanupDisk()
}

type operation struct {
	Name             string
	Method           func() error
	DiskExpected     bool
	InstanceExpected bool
	State            state.State
	Arguments        []interface{}
}

func TestBasicOperations(t *testing.T) {
	if *project == "" || *tokenPath == "" {
		t.Skip("Skipping tests because no --project or --token-path flag was passed.")
		return
	}
	ops := []operation{
		{
			Name:             "Create",
			Method:           driver.Create,
			DiskExpected:     true,
			InstanceExpected: true,
			State:            state.Running,
		},
		{
			Name:             "Stop",
			Method:           driver.Stop,
			DiskExpected:     true,
			InstanceExpected: false,
			State:            state.Stopped,
		},
		{
			Name:             "Start",
			Method:           driver.Start,
			DiskExpected:     true,
			InstanceExpected: true,
			State:            state.Running,
		},
		{
			Name:             "Restart",
			Method:           driver.Restart,
			DiskExpected:     true,
			InstanceExpected: true,
			State:            state.Running,
		},
		{
			Name:             "Remove",
			Method:           driver.Remove,
			DiskExpected:     false,
			InstanceExpected: false,
			State:            state.None,
		},
	}
	defer cleanup()
	for _, op := range ops {
		log.Info("Executing operation: ", op.Name)
		err := op.Method()
		if err != nil {
			t.Fatal(err)
		}
		AssertDiskAndInstance(op.DiskExpected, op.InstanceExpected)
		if s, _ := driver.GetState(); s != op.State {
			t.Fatalf("State should be %v, but is: %v", op.State, s)
		}
	}
}

func AssertDiskAndInstance(diskShouldExist, instShouldExist bool) {
	d, err := c.service.Disks.Get(*project, zone, "test-instance-disk").Do()
	if diskShouldExist {
		if d == nil || err != nil {
			log.Fatal("Error retrieving disk that should exist.")
		}
	} else if d != nil {
		log.Fatal("Disk shouldn't exist but does.")
	}
	i, err := c.service.Instances.Get(*project, zone, "test-instance").Do()
	if instShouldExist {
		if i == nil || err != nil {
			log.Fatal("error retrieving instance that should exist.")
		}
	} else if i != nil {
		log.Fatal("Instance shouldnt exist but does.")
	}
}

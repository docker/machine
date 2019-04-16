package yandex

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
)

// Driver is a struct compatible with the docker.hosts.drivers.Driver interface.
type Driver struct {
	*drivers.BaseDriver
	//Address string
	//Tags    string

	Token    string
	Endpoint string

	InstanceID       string
	FolderID         string
	SubnetID         string
	Zone             string
	SSHUser          string
	Cores            int
	Memory           int
	DiskSize         int
	DiskType         string
	UserDataFile     string
	PlatformID       string
	ServiceAccountID string
	UseIPv6          bool
	UseInternalIP    bool
	ImageID          string
	ImageFamilyName  string
	ImageFolderID    string
	Preemptible      bool
}

const (
	defaultSSHPort         = 22
	defaultSSHUser         = "ubuntu"
	defaultImageFamilyName = "ubuntu-1604-lts"
	defaultZone            = "ru-central1-a"
	defaultCores           = 1
	defaultMemory          = 1 * 1024 * 1024 * 1024
	defaultDiskSize        = 20 * 1024 * 1024 * 1024
	defaultDiskType        = "network-hdd"
	defaultPlatformID      = "standard-v1"
	defaultImageFolderID   = StandardImagesFolderID
)

// NewDriver creates a Driver with the specified storePath.
func NewDriver(machineName string, storePath string) *Driver {
	return &Driver{
		ImageID:       defaultImageFamilyName,
		Cores:         defaultCores,
		Memory:        defaultMemory,
		DiskSize:      defaultDiskSize,
		DiskType:      defaultDiskType,
		PlatformID:    defaultPlatformID,
		ImageFolderID: defaultImageFolderID,
		Zone:          defaultZone,
		BaseDriver: &drivers.BaseDriver{
			MachineName: machineName,
			StorePath:   storePath,
		},
	}
}

// Create creates a Yandex.Cloud VM instance acting as a docker host.
func (d *Driver) Create() error {
	log.Infof("Generating SSH Key")

	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	log.Infof("Creating instance...")
	c, err := d.buildClient()
	if err != nil {
		return err
	}

	return c.createInstance(d)

}

// DriverName returns the name of the driver
func (d *Driver) DriverName() string {
	return "yandex"
}

func (d *Driver) GetCreateFlags() []mcnflag.Flag {
	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "YC_TOKEN",
			Name:   "yandex-token",
			Usage:  "Yandex.Cloud access token",
		},
		mcnflag.StringFlag{
			EnvVar: "YC_SSH_USER",
			Name:   "yandex-ssh-user",
			Usage:  "SSH username",
			Value:  defaultSSHUser,
		},
		mcnflag.IntFlag{
			EnvVar: "YC_SSH_PORT",
			Name:   "yandex-ssh-port",
			Usage:  "SSH port",
			Value:  defaultSSHPort,
		},
		mcnflag.StringFlag{
			EnvVar: "YC_IMAGE_ID",
			Name:   "yandex-image-id",
			Usage:  "Yandex.Cloud Image",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "YC_IMAGE_FAMILY_NAME",
			Name:   "yandex-image-family-name",
			Usage:  "Yandex.Cloud Image family name to search in ",
			Value:  defaultImageFamilyName,
		},
		mcnflag.StringFlag{
			EnvVar: "YC_IMAGE_FOLDER_ID",
			Name:   "yandex-image-folder-id",
			Usage:  "Yandex.Cloud Folder ID to search image by family name defined in `--yandex-image-family-name`",
			Value:  defaultImageFolderID,
		},
		mcnflag.StringFlag{
			EnvVar: "YC_FOLDER_ID",
			Name:   "yandex-folder-id",
			Usage:  "Folder ID",
		},
		mcnflag.StringFlag{
			EnvVar: "YC_SUBNET_ID",
			Name:   "yandex-subnet-id",
			Usage:  "Subnet ID",
		},
		mcnflag.StringFlag{
			EnvVar: "YC_ZONE",
			Name:   "yandex-zone",
			Usage:  "Yandex.Cloud zone",
			Value:  defaultZone,
		},
		mcnflag.IntFlag{
			EnvVar: "YC_CORES",
			Name:   "yandex-cores",
			Usage:  "vCPUs",
			Value:  defaultCores,
		},
		mcnflag.IntFlag{
			EnvVar: "YC_MEMORY",
			Name:   "yandex-memory",
			Usage:  "Memory in bytes",
			Value:  defaultMemory,
		},
		mcnflag.IntFlag{
			EnvVar: "YC_DISK_SIZE",
			Name:   "yandex-disk-size",
			Usage:  "Disk size in bytes",
			Value:  defaultDiskSize,
		},
		mcnflag.StringFlag{
			EnvVar: "YC_DISK_TYPE",
			Name:   "yandex-disk-type",
			Usage:  "Disk type, e.g. 'network-hdd'",
			Value:  defaultDiskType,
		},
	}

}

//func (d *Driver) GetIP() (string, error) {

// GetMachineName returns the name of the machine
//func (d *Driver) GetMachineName() string { panic("implement me") }

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

// GetSSHKeyPath returns key path for use with ssh
//func (d *Driver) GetSSHKeyPath() string { panic("implement me") }

// GetSSHPort returns port for use with ssh
//func (d *Driver) GetSSHPort() (int, error) {
//	panic("implement me")
//}

// GetSSHUsername returns username for use with ssh
func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "docker-user"
	}
	return d.SSHUser
}

// GetURL returns the URL of the remote docker daemon.
func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("tcp://%s", net.JoinHostPort(ip, "2376")), nil
}

func (d *Driver) GetState() (state.State, error) {
	c, err := d.buildClient()
	if err != nil {
		return state.None, err
	}

	instance, err := c.sdk.Compute().Instance().Get(context.TODO(), &compute.GetInstanceRequest{
		InstanceId: d.InstanceID,
	})
	if err != nil {
		return state.Error, err
	}

	status := instance.Status.String()
	log.Debugf("Instance State: %s", status)

	switch status {
	case "PROVISIONING", "STARTING":
		return state.Starting, nil
	case "RUNNING":
		return state.Running, nil
	case "STOPPING", "STOPPED", "DELETING":
		return state.Stopped, nil
	}

	return state.None, nil
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) PreCreateCheck() error {
	c, err := d.buildClient()
	if err != nil {
		return err
	}

	log.Infof("Check that the folder exists")

	_, err = c.sdk.ResourceManager().Folder().Get(context.TODO(), &resourcemanager.GetFolderRequest{
		FolderId: d.FolderID,
	})
	if err != nil {
		return fmt.Errorf("Folder with ID %q not found. %v", d.FolderID, err)
	}

	resp, err := c.sdk.Compute().Instance().List(context.TODO(), &compute.ListInstancesRequest{
		FolderId: d.FolderID,
		Filter:   fmt.Sprintf("name = \"%s\"", d.MachineName),
	})
	if err != nil {
		return fmt.Errorf("Fail to get instance list in Folder: %s", err)
	}
	if len(resp.Instances) > 0 {
		return fmt.Errorf("instance %q already exists in folder %q", d.MachineName, d.FolderID)
	}

	return nil

}

func (d *Driver) Remove() error {
	c, err := d.buildClient()
	if err != nil {
		return err
	}

	ctx := context.TODO()
	op, err := c.sdk.WrapOperation(c.sdk.Compute().Instance().Delete(ctx, &compute.DeleteInstanceRequest{
		InstanceId: d.InstanceID,
	}))
	if err != nil {
		return err
	}

	return op.Wait(ctx)
}

func (d *Driver) Restart() error {
	c, err := d.buildClient()
	if err != nil {
		return err
	}

	ctx := context.TODO()
	op, err := c.sdk.WrapOperation(c.sdk.Compute().Instance().Restart(ctx, &compute.RestartInstanceRequest{
		InstanceId: d.InstanceID,
	}))
	if err != nil {
		return err
	}

	return op.Wait(ctx)
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.FolderID = flags.String("yandex-folder-id")
	if d.FolderID == "" {
		return errors.New("no Yandex.Cloud Folder ID specified (--yandex-folder-id)")
	}

	d.Token = flags.String("yandex-token")
	if d.Token == "" {
		return fmt.Errorf("Yandex.Cloud driver requires the --yandex-token option")
	}

	d.ImageID = flags.String("yandex-image-id")
	d.Zone = flags.String("yandex-zone")
	d.SSHUser = flags.String("yandex-ssh-user")
	d.SubnetID = flags.String("yandex-subnet-id")
	d.Cores = flags.Int("yandex-cores")
	d.Memory = flags.Int("yandex-memory")
	d.DiskSize = flags.Int("yandex-disk-size")
	d.DiskType = flags.String("yandex-disk-type")
	d.ImageFolderID = flags.String("yandex-image-folder-id")
	d.ImageFamilyName = flags.String("yandex-image-family-name")

	d.SetSwarmConfigFromFlags(flags)

	return nil
}

func (d *Driver) Start() error {
	c, err := d.buildClient()
	if err != nil {
		return err
	}

	ctx := context.TODO()
	op, err := c.sdk.WrapOperation(c.sdk.Compute().Instance().Start(ctx, &compute.StartInstanceRequest{
		InstanceId: d.InstanceID,
	}))
	if err != nil {
		return err
	}

	return op.Wait(ctx)
}

func (d *Driver) Stop() error {
	c, err := d.buildClient()
	if err != nil {
		return err
	}

	ctx := context.TODO()
	op, err := c.sdk.WrapOperation(c.sdk.Compute().Instance().Stop(ctx, &compute.StopInstanceRequest{
		InstanceId: d.InstanceID,
	}))
	if err != nil {
		return err
	}

	return op.Wait(ctx)
}

func (d *Driver) buildClient() (*YandexCloudClient, error) {
	return NewYandexCloudClient(d)
}

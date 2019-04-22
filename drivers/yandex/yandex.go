package yandex

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"text/template"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/resourcemanager/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
)

// Driver is a struct compatible with the docker.hosts.drivers.Driver interface.
type Driver struct {
	*drivers.BaseDriver

	Endpoint              string
	ServiceAccountKeyFile string
	Token                 string

	CloudID         string
	Cores           int
	DiskSize        int
	DiskType        string
	FolderID        string
	ImageFamilyName string
	ImageFolderID   string
	ImageID         string
	InstanceID      string
	Labels          []string
	Memory          int
	PlatformID      string
	Preemptible     bool
	SSHUser         string
	SubnetID        string
	UseIPv6         bool
	UseInternalIP   bool
	UserDataFile    string
	UserData        string
	Zone            string
}

const (
	defaultCores           = 1
	defaultDiskSize        = 20
	defaultDiskType        = "network-hdd"
	defaultEndpoint        = "api.cloud.yandex.net:443"
	defaultImageFamilyName = "ubuntu-1604-lts"
	defaultImageFolderID   = StandardImagesFolderID
	defaultMemory          = 1
	defaultPlatformID      = "standard-v1"
	defaultSSHPort         = 22
	defaultSSHUser         = "yc-user"
	defaultZone            = "ru-central1-a"
)

// NewDriver creates a Driver with the specified storePath.
func NewDriver(machineName string, storePath string) *Driver {
	return &Driver{
		Cores:         defaultCores,
		DiskSize:      defaultDiskSize,
		DiskType:      defaultDiskType,
		ImageFolderID: defaultImageFolderID,
		ImageID:       defaultImageFamilyName,
		Memory:        defaultMemory,
		PlatformID:    defaultPlatformID,
		Zone:          defaultZone,
		BaseDriver: &drivers.BaseDriver{
			MachineName: machineName,
			StorePath:   storePath,
		},
	}
}

// Create creates a Yandex.Cloud VM instance acting as a docker host.
func (d *Driver) Create() error {
	log.Infof("Prepare instance user-data")
	if err := d.prepareUserData(); err != nil {
		return err
	}
	log.Debugf("Formed user-data:\n%s\n", d.UserData)

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
			Usage:  "Yandex.Cloud OAuth token",
		},
		mcnflag.StringFlag{
			EnvVar: "YC_SERVICE_ACCOUNT_KEY_FILE",
			Name:   "yandex-service-account-key-file",
			Usage:  "Yandex.Cloud Service Account key file",
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
			Usage:  "Yandex.Cloud Image identifier",
			Value:  "",
		},
		mcnflag.StringFlag{
			EnvVar: "YC_IMAGE_FAMILY_NAME",
			Name:   "yandex-image-family-name",
			Usage:  "Yandex.Cloud Image family name to lookup image ID for instance",
			Value:  defaultImageFamilyName,
		},
		mcnflag.StringFlag{
			EnvVar: "YC_IMAGE_FOLDER_ID",
			Name:   "yandex-image-folder-id",
			Usage:  "Yandex.Cloud Folder ID to latest image by family name defined in `--yandex-image-family-name`",
			Value:  defaultImageFolderID,
		},
		mcnflag.StringFlag{
			EnvVar: "YC_ENDPOINT",
			Name:   "yandex-endpoint",
			Usage:  "Yandex.Cloud API Endpoint",
			Value:  defaultEndpoint,
		},
		mcnflag.StringFlag{
			EnvVar: "YC_FOLDER_ID",
			Name:   "yandex-folder-id",
			Usage:  "Folder ID",
		},
		mcnflag.StringFlag{
			EnvVar: "YC_CLOUD_ID",
			Name:   "yandex-cloud-id",
			Usage:  "Cloud ID",
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
			Usage:  "Count of virtual CPUs",
			Value:  defaultCores,
		},
		mcnflag.IntFlag{
			EnvVar: "YC_MEMORY",
			Name:   "yandex-memory",
			Usage:  "Memory in gigabytes",
			Value:  defaultMemory,
		},
		mcnflag.IntFlag{
			EnvVar: "YC_DISK_SIZE",
			Name:   "yandex-disk-size",
			Usage:  "Disk size in gigabytes",
			Value:  defaultDiskSize,
		},
		mcnflag.StringFlag{
			EnvVar: "YC_DISK_TYPE",
			Name:   "yandex-disk-type",
			Usage:  "Disk type, e.g. 'network-hdd'",
			Value:  defaultDiskType,
		},
		mcnflag.StringSliceFlag{
			EnvVar: "YC_LABELS",
			Name:   "yandex-labels",
			Usage:  "Instance labels in 'key=value' format",
		},
		mcnflag.BoolFlag{
			EnvVar: "YC_PREEMPTIBLE",
			Name:   "yandex-preemptible",
			Usage:  "Yandex.Cloud Instance preemptibility flag",
		},
		mcnflag.BoolFlag{
			EnvVar: "YANDEX_USE_INTERNAL_IP",
			Name:   "yandex-use-internal-ip",
			Usage:  "Use internal Instance IP rather than public one",
		},
		mcnflag.StringFlag{
			EnvVar: "YC_USERDATA",
			Name:   "yandex-userdata",
			Usage:  "Path to file with cloud-init user-data",
		},
	}

}

// GetSSHHostname returns hostname for use with ssh
func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

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

	status := instance.Status
	log.Debugf("Instance State: %s", status)

	switch status {
	case compute.Instance_PROVISIONING, compute.Instance_STARTING:
		return state.Starting, nil
	case compute.Instance_RUNNING:
		return state.Running, nil
	case compute.Instance_STOPPING, compute.Instance_STOPPED, compute.Instance_DELETING:
		return state.Stopped, nil
	}

	return state.None, nil
}

func (d *Driver) Kill() error {
	return d.Stop()
}

func (d *Driver) PreCreateCheck() error {
	//d.UserData = defaultUserData(d.SSHKeyPath)
	if d.UserDataFile != "" {
		if _, err := os.Stat(d.UserDataFile); os.IsNotExist(err) {
			return fmt.Errorf("user-data file %s could not be found", d.UserDataFile)
		}
	}

	c, err := d.buildClient()
	if err != nil {
		return err
	}

	if d.FolderID == "" {
		if d.CloudID == "" {
			log.Warn("No Folder and Cloud identifiers provided")
			log.Warn("Try guess cloud ID to use")
			d.CloudID, err = d.guessCloudID()
			if err != nil {
				return err
			}
		}

		log.Warnf("Try guess folder ID to use inside cloud %q", d.CloudID)
		d.FolderID, err = d.guessFolderID()
		if err != nil {
			return err
		}
	}
	log.Infof("Check that folder exists")
	folder, err := c.sdk.ResourceManager().Folder().Get(context.TODO(), &resourcemanager.GetFolderRequest{
		FolderId: d.FolderID,
	})
	if err != nil {
		return fmt.Errorf("Folder with ID %q not found. %v", d.FolderID, err)
	}

	log.Infof("Check if the instance with name %q already exists in folder", d.MachineName)
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

	if d.SubnetID == "" {
		log.Warnf("Subnet ID not provided, will search one for Zone %q in folder %q [%s]", d.Zone, folder.Name, folder.Id)
		d.SubnetID, err = d.findSubnetID()
		if err != nil {
			return err
		}

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

	d.CloudID = flags.String("yandex-cloud-id")
	d.Cores = flags.Int("yandex-cores")
	d.DiskSize = flags.Int("yandex-disk-size")
	d.DiskType = flags.String("yandex-disk-type")
	d.Endpoint = flags.String("yandex-endpoint")
	d.ImageFamilyName = flags.String("yandex-image-family-name")
	d.ImageFolderID = flags.String("yandex-image-folder-id")
	d.ImageID = flags.String("yandex-image-id")
	d.Labels = flags.StringSlice("yandex-labels")
	d.Memory = flags.Int("yandex-memory")
	d.Preemptible = flags.Bool("yandex-preemptible")
	d.SSHUser = flags.String("yandex-ssh-user")
	d.SSHPort = flags.Int("yandex-ssh-port")
	d.SubnetID = flags.String("yandex-subnet-id")
	d.UseInternalIP = flags.Bool("yandex-use-internal-ip")
	d.UserDataFile = flags.String("yandex-userdata")
	d.Zone = flags.String("yandex-zone")

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

func (d *Driver) buildClient() (*YCClient, error) {
	return NewYCClient(d)
}

func (d *Driver) guessCloudID() (string, error) {
	c, err := d.buildClient()
	if err != nil {
		return "", err
	}

	resp, err := c.sdk.ResourceManager().Cloud().List(context.TODO(), &resourcemanager.ListCloudsRequest{})
	if err != nil {
		return "", err
	}
	switch {
	case len(resp.Clouds) == 0:
		return "", errors.New("no one Cloud available")
	case len(resp.Clouds) > 1:
		return "", errors.New("more than one Cloud available, could not choose one")
	}
	return resp.Clouds[0].Id, nil
}

func (d *Driver) guessFolderID() (string, error) {
	c, err := d.buildClient()
	if err != nil {
		return "", err
	}

	resp, err := c.sdk.ResourceManager().Folder().List(context.TODO(), &resourcemanager.ListFoldersRequest{
		CloudId: d.CloudID,
	})
	if err != nil {
		return "", err
	}
	switch {
	case len(resp.Folders) == 0:
		return "", errors.New("no one Folder available")
	case len(resp.Folders) > 1:
		return "", errors.New("more than one Folder available, could not choose one")
	}

	return resp.Folders[0].Id, nil
}

func (d *Driver) findSubnetID() (string, error) {
	c, err := d.buildClient()
	if err != nil {
		return "", err
	}

	ctx := context.TODO()

	resp, err := c.sdk.VPC().Subnet().List(ctx, &vpc.ListSubnetsRequest{
		FolderId: d.FolderID,
	})
	if err != nil {
		return "", err
	}

	for _, subnet := range resp.Subnets {
		if subnet.ZoneId != d.Zone {
			continue
		}
		return subnet.Id, nil
	}
	return "", fmt.Errorf("no subnets in zone: %s", d.Zone)
}

func (d *Driver) ParsedLabels() map[string]string {
	var labels = make(map[string]string)

	for _, labelPair := range d.Labels {
		labelPair = strings.TrimSpace(labelPair)
		chunks := strings.SplitN(labelPair, "=", 2)
		if len(chunks) == 1 {
			labels[chunks[0]] = ""
		} else {
			labels[chunks[0]] = chunks[1]
		}
	}
	return labels
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) prepareUserData() error {
	if d.UserDataFile != "" {
		log.Infof("Use provided file %q with user-data", d.UserDataFile)
		buf, err := ioutil.ReadFile(d.UserDataFile)
		if err != nil {
			return err
		}
		d.UserData = string(buf)
		return nil
	}

	log.Infof("Generating SSH Key")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}

	d.UserData, err = defaultUserData(d.GetSSHUsername(), string(publicKey))

	return err
}

func defaultUserData(sshUserName, sshPublicKey string) (string, error) {
	type templateData struct {
		SSHUserName  string
		SSHPublicKey string
	}
	buf := &bytes.Buffer{}
	err := defaultUserDataTemplate.Execute(buf, templateData{
		SSHUserName:  sshUserName,
		SSHPublicKey: sshPublicKey,
	})
	if err != nil {
		return "", fmt.Errorf("error while process template: %s", err)
	}

	return buf.String(), nil
}

var defaultUserDataTemplate = template.Must(
	template.New("user-data").Parse(`#cloud-config
ssh_pwauth: no

users:
  - name: {{.SSHUserName}}
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
      - {{.SSHPublicKey}}
`))

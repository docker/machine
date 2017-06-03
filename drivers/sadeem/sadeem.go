package sadeem

import (
	"fmt"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/ssh"
	"github.com/docker/machine/libmachine/state"
	"io/ioutil"
	"time"
)

type vmData struct {
	Hostname          string
	OfferId           string
	TemplateId        string
	DatacenterId      string
	PrivateNetworking bool
	DiskSize          int
	CpuCount          int
	VCpu              int
	Memory            int
}

type Driver struct {
	*drivers.BaseDriver
	ApiKey   string
	ClientId string
	VmId     string
	VmData   *vmData
	SSHKey   string
}

const (
	defaultSSHPort        = 22
	defaultSSHUser        = "root"
	defaultPrivateNetwork = false
	defaultDatacenter     = "riyadh"
	defaultOffer          = "25"
	defaultTemplate       = "ubuntu"
)

func (d *Driver) GetCreateFlags() []mcnflag.Flag {

	return []mcnflag.Flag{
		mcnflag.StringFlag{
			EnvVar: "SADEEM_APIKEY",
			Name:   "sadeem-apikey",
			Usage:  "Sadeem Api Key",
		},
		mcnflag.StringFlag{
			EnvVar: "SADEEM_CLIENT_ID",
			Name:   "sadeem_client_id",
			Usage:  "Sadeem Client Id",
		},
		mcnflag.StringFlag{
			EnvVar: "SADEEM_DATACENTER",
			Name:   "sadeem_datacenter",
			Usage:  "Sadeem Datacenter Name",
			Value:  defaultDatacenter,
		},
		mcnflag.StringFlag{
			EnvVar: "SADEEM_SERVICE_OFFER",
			Name:   "sadeem_service_offer",
			Usage:  "Sadeem Service Offer Name",
			Value:  defaultOffer,
		},
		mcnflag.StringFlag{
			EnvVar: "SADEEM_TEMPLATE",
			Name:   "sadeem_template",
			Usage:  "Sadeem Template Name",
			Value:  defaultTemplate,
		},
		mcnflag.StringFlag{
			EnvVar: "SADEEM_PRIVATE_NETWORK",
			Name:   "sadeem_private_network",
			Usage:  "enable private networking for Vm",
		},
	}
}

func ValidateDriverConfig(d *Driver) error {
	if d.ApiKey == "" {
		return fmt.Errorf("sadeem driver requires the --sadeem-apikey")
	}

	if d.ClientId == "" {
		return fmt.Errorf("sadeem driver requires the --sadeem_client_id")
	}
	return nil
}

func NewDriver(hostName, storePath string) *Driver {

	return &Driver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: hostName,
			StorePath:   storePath,
		},
		VmData: &vmData{
			Hostname: hostName,
		},
	}
}

func (d *Driver) DriverName() string {
	return "sadeem"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.ApiKey = flags.String("sadeem-apikey")
	d.ClientId = flags.String("sadeem_client_id")
	c := NewClient(d.ApiKey, d.ClientId)

	DCName := flags.String("sadeem_datacenter")
	Template := flags.String("sadeem_template")
	OfferName := flags.String("sadeem_service_offer")

	DcId, e := c.GetDatacenterId(DCName)

	if e != nil {
		return e
	}

	TempId, e := c.GetTemplateId(Template)
	if e != nil {
		return e
	}

	OfferId, e := c.GetOferrId(OfferName, DcId)

	if e != nil {
		return e
	}

	d.VmData = &vmData{
		OfferId:           OfferId,
		TemplateId:        TempId,
		DatacenterId:      DcId,
		PrivateNetworking: false,
	}

	d.SetSwarmConfigFromFlags(flags)
	return ValidateDriverConfig(d)

}

func (d *Driver) GetURL() (string, error) {
	if err := drivers.MustBeRunning(d); err != nil {
		return "", err
	}

	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHPort() (int, error) {
	return 22, nil
}

func (d *Driver) GetSSHUsername() string {
	return "root"
}

func (d *Driver) Create() error {
	key, err := d.CreateSSHKey()

	if err != nil {
		return err
	}

	ApiKey := d.ApiKey
	ClientId := d.ClientId

	c := NewClient(ApiKey, ClientId)
	d.SSHKey = key

	offer_id := d.VmData.OfferId
	tmp_id := d.VmData.TemplateId
	dc_id := d.VmData.DatacenterId

	vmId, err := c.CreateNewVM(d.MachineName, dc_id, offer_id, tmp_id, key)
	if err != nil {
		return err
	}
	IP, err := c.GetVmIP(vmId)

	if err != nil {
		return err
	}

	d.VmId = vmId
	d.IPAddress = string(IP)

	for {

		b, _ := c.GetVmState(d.VmId)

		switch b {
		case "pending":
			break
		case "archived":
			return fmt.Errorf("Failed to create VM")
		default:
			return nil
		}

		time.Sleep(1 * time.Second)
	}
	return nil
}

func (d *Driver) Start() error {

	ApiKey := d.ApiKey
	ClientId := d.ClientId
	client := NewClient(ApiKey, ClientId)
	err := client.StartVm(d.VmId)

	return err
}

func (d *Driver) Stop() error {
	ApiKey := d.ApiKey
	ClientId := d.ClientId
	client := NewClient(ApiKey, ClientId)
	err := client.StopVm(d.VmId)

	return err
}

func (d *Driver) Restart() error {
	ApiKey := d.ApiKey
	ClientId := d.ClientId
	client := NewClient(ApiKey, ClientId)
	err := client.RebootVm(d.VmId)

	return err
}

func (d *Driver) Kill() error {
	ApiKey := d.ApiKey
	ClientId := d.ClientId
	client := NewClient(ApiKey, ClientId)
	err := client.KillVm(d.VmId)
	return err
}

func (d *Driver) Remove() error {

	ApiKey := d.ApiKey
	ClientId := d.ClientId
	client := NewClient(ApiKey, ClientId)
	err := client.destroyVm(d.VmId)
	return err
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) CreateSSHKey() (string, error) {
	SSHKeyPath := d.GetSSHKeyPath()

	if err := ssh.GenerateSSHKey(SSHKeyPath); err != nil {
		return "", err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return "", err
	}

	return string(publicKey), err
}

func (d *Driver) GetState() (state.State, error) {
	ApiKey := d.ApiKey
	ClientId := d.ClientId
	client := NewClient(ApiKey, ClientId)
	b, err := client.GetVmState(d.VmId)

	if err != nil {
		return state.Error, err
	}

	switch b {
	case "pending":
		return state.Starting, nil
	case "running":
		return state.Running, nil
	case "power-off":
		return state.Stopped, nil
	case "archived":
		return state.Error, fmt.Errorf("Vm already deleted")
	}
	return state.None, nil
}

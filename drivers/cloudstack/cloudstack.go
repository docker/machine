package cloudstack

import (
    "fmt"
    "os/exec"

    "github.com/docker/machine/drivers"
    "github.com/docker/machine/state"
    log "github.com/Sirupsen/logrus"
    gcs "github.com/mindjiver/gopherstack"
    flag "github.com/docker/docker/pkg/mflag"
    
)

type Driver struct {
    MachineName         string
    ApiURL              string
    ApiKey              string
    SecretKey           string
    TemplateName        string
    TemplateId          string
    OfferId             string
    ZoneId              string
    storePath           string
}

type CreateFlags struct {
    ApiURL              *string
    ApiKey              *string
    SecretKey           *string
    TemplateName        *string
    OfferId             *string
    ZoneId              *string
}

func init() {
    drivers.Register("cloudstack", &drivers.RegisteredDriver{
        New:                 NewDriver,
        RegisterCreateFlags: RegisterCreateFlags,
    })
}

// RegisterCreateFlags registers the flags this driver adds to
// "machine create"
func RegisterCreateFlags(cmd *flag.FlagSet) interface{} {
    createFlags := new(CreateFlags)
    createFlags.ApiURL = cmd.String(
        []string{"-cloudstack-api-url"},
        "",
        "Your cloudstack API URL",
    )
    createFlags.ApiKey = cmd.String(
        []string{"-cloudstack-api-key"},
        "",
        "Your cloudstack API key",
    )
    createFlags.SecretKey = cmd.String(
        []string{"-cloudstack-secret-key"},
        "",
        "Your cloudstack secret key",
    )
    createFlags.OfferId = cmd.String(
        []string{"-cloudstack-offer-id"},
        "",
        "Your cloudstack offer's ID ",
    )
    createFlags.TemplateName = cmd.String(
        []string{"-cloudstack-template"},
        "",
        "Your cloudstack template name to use",
    )
    createFlags.TemplateName = cmd.String(
        []string{"-cloudstack-zone-id"},
        "",
        "Your cloudstack zone id to use",
    )

    return createFlags
}

func NewDriver(storePath string) (drivers.Driver, error) {
    return &Driver{storePath: storePath}, nil
}

func (d *Driver) DriverName() string {
    return "cloudstack"
}

func (d *Driver) SetConfigFromFlags(flagsInterface interface{}) error {
    flags := flagsInterface.(*CreateFlags)

    d.ApiURL = *flags.ApiURL
    if d.ApiURL == "" {
        return fmt.Errorf("cloudstack driver requires the --cloudstack-api-url option")
    }

    d.ApiKey = *flags.ApiKey
    if d.ApiKey == "" {
        return fmt.Errorf("cloudstack driver requires the --cloudstack-api-key option")
    }

    d.SecretKey = *flags.SecretKey
    if d.SecretKey == "" {
        return fmt.Errorf("cloudstack driver requires the --cloudstack-secret-key option")
    }

    d.OfferId = *flags.OfferId
    if d.OfferId == "" {
        return fmt.Errorf("cloudstack driver requires the --cloudstack-offer-id option")
    }

    d.TemplateName = *flags.TemplateName
    if d.TemplateName == "" {
        return fmt.Errorf("cloudstack driver requires the --cloudstack-template option")
    }

    d.ZoneId = *flags.ZoneId
    if d.ZoneId == "" {
        return fmt.Errorf("cloudstack driver requires the --cloudstack-zone-id option")
    }

    return nil
}

func (d *Driver) Create() error {

    log.Infof("Creating Cloudstack instance...")

    client := d.getClient()

    /** First we have to fetch some IDs before creating the instance **/
    log.Infof("Fetching template id from provided template name : %q",
        d.TemplateName,
    )
    responseTemplateList, err := client.ListTemplates(d.TemplateName,"executable")
    if err != nil {
        return err
    }
    d.TemplateId = responseTemplateList.Listtemplatesresponse.Template[0].ID
    if len(responseTemplateList.Listtemplatesresponse.Template) > 1 {
        log.Infof("More than one template has been found with the name %q. I'm gonna use the first with the ID %q",
            d.TemplateName,
            d.TemplateId
        )
    }
    d.TemplateName = responseTemplateList.Listtemplatesresponse.Template[0].Name


    // TODO : implement listServiceOfferings CS's API into gopherstack and use offer name instead ID (more user friendly)
    log.Infof("Offer ID is %q",
        d.OfferId,
    )

    // TODO : implement listZones CS's API into gopherstack and use zone name instead ID (more user friendly)
    log.Infof("Zone ID is %q",
        d.ZoneId,
    )

    d.setMachineNameIfNotSet()

    log.Infof("Creating SSH key...")

    if err := ssh.GenerateSSHKey(path.Join(d.storePath, "id_rsa")); err != nil {
        return err
    }

    return nil
}

func (d *Driver) Start() error {
    return fmt.Errorf("Not implemented yet")
}

func (d *Driver) GetURL() (string, error) {
    return "foo", nil
}

func (d *Driver) GetIP() (string, error) {
    return "", nil
}

func (d *Driver) GetState() (state.State, error) {
    return state.None, nil
}

func (d *Driver) Stop    () error {
    return fmt.Errorf("hosts without a driver cannot be stopped")
}

func (d *Driver) Remove() error {
    return nil
}

func (d *Driver) Restart() error {
    return fmt.Errorf("hosts without a driver cannot be restarted")
}

func (d *Driver) Kill() error {
    return fmt.Errorf("hosts without a driver cannot be killed")
}

func (d *Driver) Upgrade() error {
    return fmt.Errorf("hosts without a driver cannot be upgraded")
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
    return nil, fmt.Errorf("hosts without a driver do not support SSH")
}

func (d *Driver) getClient() *gcs.CloudstackClient {

    client := gcs.CloudstackClient{}.New(d.ApiURL, d.ApiKey,
        d.SecretKey, true)

    return client
}

func (d *Driver) setMachineNameIfNotSet() {
    if d.MachineName == "" {
        d.MachineName = fmt.Sprintf("docker-host-cloudstack-%s", utils.GenerateRandomID())
    }
}

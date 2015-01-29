package centurylinkcloud

import (
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/CenturyLinkLabs/clcgo"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/state"
)

const statusWaitSeconds = 10

type Driver struct {
	MachineName    string
	CaCertPath     string
	PrivateKeyPath string
	storePath      string
	BearerToken    string
	AccountAlias   string
	ServerID       string
	Username       string
	Password       string
	GroupID        string
	SourceServerID string
	CPU            int
	MemoryGB       int
}

func init() {
	drivers.Register("centurylinkcloud", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: getCreateFlags,
	})
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	return &Driver{MachineName: machineName, storePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey}, nil
}

func (d *Driver) AuthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) DeauthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) GetSSHPort() (int, error) {
	return 22, nil
}

func (d *Driver) GetSSHUsername() string {
	return "root"
}

func (d *Driver) DriverName() string {
	return "centurylinkcloud"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Username = flags.String("centurylinkcloud-username")
	d.Password = flags.String("centurylinkcloud-password")
	d.GroupID = flags.String("centurylinkcloud-group-id")
	d.SourceServerID = flags.String("centurylinkcloud-source-server-id")
	d.CPU = flags.Int("centurylinkcloud-cpu")
	d.MemoryGB = flags.Int("centurylinkcloud-memory-gb")

	if d.Username == "" {
		return fmt.Errorf("centurylinkcloud driver requires the --centurylinkcloud-username option")
	}

	if d.GroupID == "" {
		return fmt.Errorf("centurylinkcloud driver requires the --centurylinkcloud-group-id option")
	}

	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	_, s, err := d.getServer()
	if err != nil {
		return "", err
	}

	address := publicIPFromServer(s)
	if address != "" {
		return address, nil
	}

	return "", errors.New("no IP could be found for this server")
}

func (d *Driver) GetState() (state.State, error) {
	_, s, err := d.getServer()
	if err != nil {
		return state.Error, err
	}

	if s.IsActive() {
		return state.Running, nil
	} else if s.IsPaused() {
		return state.Paused, nil
	}

	return state.Stopped, nil
}

func (d *Driver) Remove() error {
	c, s, err := d.getServer()
	if err != nil {
		return err
	}

	st, err := c.DeleteEntity(&s)
	if err != nil {
		return err
	}

	for !st.HasSucceeded() {
		time.Sleep(time.Second * statusWaitSeconds)
		if err := c.GetEntity(&st); err != nil {
			return err
		}
		log.Debugf("Deletion status: %s", st.Status)
	}

	return nil
}

func (d *Driver) Start() error {
	if err := d.doOperation(clcgo.PowerOnServer); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Stop() error {
	if err := d.doOperation(clcgo.PowerOffServer); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Restart() error {
	if err := d.doOperation(clcgo.RebootServer); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Kill() error {
	if err := d.doOperation(clcgo.PowerOffServer); err != nil {
		return err
	}

	return nil
}

func (d *Driver) getClientWithPersistence() (*clcgo.Client, error) {
	c := clcgo.NewClient()
	if d.BearerToken == "" || d.AccountAlias == "" {
		if err := d.updateAPICredentials(c); err != nil {
			return nil, err
		}
	} else {
		c.APICredentials = clcgo.APICredentials{
			BearerToken:  d.BearerToken,
			AccountAlias: d.AccountAlias,
		}

		// Something to validate your BearerToken.
		err := c.GetEntity(&clcgo.DataCenters{})
		if err != nil {
			if rerr, ok := err.(clcgo.RequestError); ok && rerr.StatusCode == 401 {
				err := d.updateAPICredentials(c)
				if err != nil {
					return c, err
				}

				return c, nil
			}

			return c, err
		}
	}

	return c, nil
}

func (d *Driver) updateAPICredentials(c *clcgo.Client) error {
	if err := c.GetAPICredentials(d.Username, d.Password); err != nil {
		return err
	}
	// Known Issue: the store is not saved after initial machine create, and so
	// upon BearerToken expiration it will be making one token request for every
	// machine command for the rest of time, as it can't persist the token it
	// fetches.
	d.AccountAlias = c.APICredentials.AccountAlias
	d.BearerToken = c.APICredentials.BearerToken

	return nil
}

func (d *Driver) getServer() (*clcgo.Client, clcgo.Server, error) {
	s := clcgo.Server{ID: d.ServerID}
	c, err := d.getClientWithPersistence()
	if err != nil {
		return nil, s, err
	}

	err = c.GetEntity(&s)

	if err != nil {
		if rerr, ok := err.(clcgo.RequestError); ok {
			if rerr.StatusCode == 404 {
				return nil, s, fmt.Errorf("unable to find a server with the ID '%s'", d.ServerID)
			}
		}

		return nil, s, err
	}

	return c, s, nil
}

func (d *Driver) doOperation(t clcgo.OperationType) error {
	c, s, err := d.getServer()
	if err != nil {
		return err
	}

	log.Infof("Performing '%s' operation on '%s'...", t, s.ID)
	o := clcgo.ServerOperation{Server: s, OperationType: t}
	st, err := c.SaveEntity(&o)
	if err != nil {
		return nil
	}

	for !st.HasSucceeded() {
		time.Sleep(time.Second * statusWaitSeconds)
		if err := c.GetEntity(&st); err != nil {
			return err
		}
		log.Debugf("Operation status: %s", st.Status)
	}

	return nil
}

func logAndReturnError(err error) error {
	if rerr, ok := err.(clcgo.RequestError); ok {
		for f, ms := range rerr.Errors {
			for _, m := range ms {
				log.Errorf("%v: %v", f, m)
			}
		}

		return rerr
	}

	return err
}

func publicIPFromServer(s clcgo.Server) string {
	addresses := s.Details.IPAddresses
	for _, a := range addresses {
		if a.Public != "" {
			return a.Public
		}
	}

	return ""
}

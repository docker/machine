package concerto

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/utils"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/drivers/concerto/types"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

type Driver struct {
	PrivateKey  string
	CaCert      string
	ID          string
	Name        string
	certificate tls.Certificate
	connection  *http.Client
	IPAddress   string
	Plan        string
	Size        string
	endpoint    string
	storePath   string
	SshIdentity string
	URL         string
}

type CreateFlags struct {
	PrivateKey  *string
	CaCert      *string
	Plan        *string
	SshIdentity *string
}

func init() {
	drivers.Register("concerto", &drivers.RegisteredDriver{
		New:                 NewDriver,
		RegisterCreateFlags: RegisterCreateFlags,
	})
}

// RegisterCreateFlags registers the flags this driver adds to
// "docker machines create"
func RegisterCreateFlags(cmd *flag.FlagSet) interface{} {
	createFlags := new(CreateFlags)
	createFlags.CaCert = cmd.String(
		[]string{"-concerto-api-ca-key"},
		fmt.Sprintf("%s/.concerto/cert.crt", os.Getenv("HOME")),
		"Concerto API CA cert",
	)
	createFlags.PrivateKey = cmd.String(
		[]string{"-concerto-api-private-key"},
		fmt.Sprintf("%s/.concerto/private/cert.key", os.Getenv("HOME")),
		"Concerto API Private Key",
	)
	createFlags.SshIdentity = cmd.String(
		[]string{"-concerto-ssh-identity"},
		"~/.ssh/id_rsa",
		"Concerto Ssh Identity",
	)
	createFlags.Plan = cmd.String(
		[]string{"-concerto-plan-id"},
		"",
		"Concerto Cloud Plan Id",
	)

	return createFlags
}

func NewDriver(storePath string) (drivers.Driver, error) {
	driver := &Driver{storePath: storePath}
	return driver, nil
}

func (driver *Driver) DriverName() string {
	return "concerto"
}

func (d *Driver) SetConfigFromFlags(flagsInterface interface{}) error {
	flags := flagsInterface.(*CreateFlags)
	d.PrivateKey = *flags.PrivateKey
	d.CaCert = *flags.CaCert
	d.Plan = *flags.Plan
	d.SshIdentity = *flags.SshIdentity

	if d.Plan == "" {
		return fmt.Errorf("concerto driver requires the --concerto-plan-id option")
	}

	d.endpoint = "https://clients.concerto.io:886"
	d.loadCertificates()
	d.createConnection()

	return nil
}

func (d *Driver) Post(url string, parameters url.Values) ([]byte, error) {
	request, err := http.NewRequest("POST", url, bytes.NewBufferString(parameters.Encode()))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Content-Length", strconv.Itoa(len(parameters.Encode())))

	response, err := d.connection.Do(request)

	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		return nil, errors.New("Not Found")
	} else {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
}

func (d *Driver) Get(url string) ([]byte, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	response, err := d.connection.Do(request)
	defer response.Body.Close()

	if err != nil {
		return nil, err
	}
	if response.StatusCode >= 400 {
		return nil, errors.New("Not Found")
	} else {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}
}

func (d *Driver) loadCertificates() {
	/**
	 * Loads Clients Certificates and creates and 509KeyPair
	 */
	var err error
	d.certificate, err = tls.LoadX509KeyPair(d.CaCert, d.PrivateKey)
	if err != nil {
		log.Warningf("Load Certificate error (%s, %s): %s", d.CaCert, d.PrivateKey, err.Error())
	}
}

func (d *Driver) createConnection() {
	/**
	 * Creates a client with specific transport configurations
	 */
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{Certificates: []tls.Certificate{d.certificate}, InsecureSkipVerify: true},
	}
	d.connection = &http.Client{Transport: transport}
}

func (d *Driver) setNameIfNotSet() {
	if d.Name == "" {
		d.Name = fmt.Sprintf("dhost-%s", utils.GenerateRandomID()[0:4])
	}
}

func (d *Driver) ListMachines(parameters url.Values) (types.Fleet, error) {
	output, err := d.Get(d.endpoint + "/krane/ships")

	if err != nil {

		return types.Fleet{}, nil
	}
	var jsonShips types.Ships
	json.Unmarshal(output, &jsonShips)

	var final []types.Ship

	if parameters.Get("state") != "" {
		for _, ship := range jsonShips.Ships {
			if parameters.Get("state") == ship.State {
				final = append(final, ship)
			}
		}
	} else {
		final = jsonShips.Ships
	}

	return types.NewFleet(final), nil
}

func (d *Driver) FindMachine(nameOrId string) types.Ship {
	fleet, _ := d.ListMachines(nil)
	return fleet.Find(nameOrId)
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:27017", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	if d.IPAddress == "" {
		return "", fmt.Errorf("IP address is not set")
	}
	return d.IPAddress, nil
}

func (d *Driver) GetState() (state.State, error) {
	return state.None, nil
}

func (d *Driver) Create() error {
	d.setNameIfNotSet()

	log.Infof("Creating Machine in Concerto")
	parameters := url.Values{}
	parameters.Set("name", d.Name)
	parameters.Set("plan", d.Plan)
	output, err := d.Post(d.endpoint+"/krane/ships", parameters)
	if err != nil {
		return errors.New("Unable to Create Ship")
	}
	var inspect_json map[string]interface{}
	err = json.Unmarshal(output, &inspect_json)
	if err != nil {
		return errors.New("Unable to marshal to json ship response")
	}

	d.ID = inspect_json["id"].(string)

	log.Infof("Obtaining Machine Ip Address for %s", d.ID)
	for {
		newShip := d.FindMachine(d.ID)

		if newShip.Ip != "0.0.0.0" {
			d.IPAddress = newShip.Ip
		}

		if d.IPAddress != "" {
			log.Infof(" + Ip address %s", d.IPAddress)
			break
		} else {
			log.Infof(" + Polling again for Ip address so far %s", newShip.Ip)
		}

		time.Sleep(15 * time.Second)
	}

	log.Infof("Waiting for Docker Machine to be operational")
	for {
		newShip := d.FindMachine(d.ID)

		if newShip.State == "operational" {
			break
		} else {
			log.Infof(" + Polling again for state currently %s", newShip.State)
		}

		time.Sleep(15 * time.Second)
	}

	log.Infof("Machine %s(%s) is fully operational. Good luck :D", d.Name, d.IPAddress)

	return nil
}

func (d *Driver) Start() error {
	return fmt.Errorf("Not implemented yet")
}

func (d *Driver) Stop() error {
	return fmt.Errorf("Not implemented yet")
}

func (d *Driver) Remove() error {
	return nil
}

func (d *Driver) Restart() error {
	return fmt.Errorf("Not implemented yet")
}

func (d *Driver) Kill() error {
	return fmt.Errorf("Not implemented yet")
}

func (d *Driver) Upgrade() error {
	return fmt.Errorf("Not implemented yet")
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return ssh.GetSSHCommand(d.IPAddress, 22, "root", d.sshKeyPath(), args...), nil
}

func (d *Driver) sshKeyPath() string {
	return path.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

package libmachine

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/docker/machine/drivers"
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/swarm"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
	"github.com/docker/machine/utils"
	"github.com/docker/machine/version"
)

var (
	validHostNameChars                = `[a-zA-Z0-9\-\.]`
	validHostNamePattern              = regexp.MustCompile(`^` + validHostNameChars + `+$`)
	errMachineMustBeRunningForUpgrade = errors.New("Error: machine must be running to upgrade.")
	stateTimeoutDuration              = time.Second * 3
)

type Host struct {
	ConfigVersion int
	Driver        drivers.Driver
	DriverName    string
	HostOptions   *HostOptions
	Name          string `json:"-"`
	StorePath     string
}

type HostOptions struct {
	Driver        string
	Memory        int
	Disk          int
	EngineOptions *engine.EngineOptions
	SwarmOptions  *swarm.SwarmOptions
	AuthOptions   *auth.AuthOptions
}

type HostMetadata struct {
	ConfigVersion int
	DriverName    string
	HostOptions   HostOptions
}

type HostListItem struct {
	Name         string
	Active       bool
	DriverName   string
	State        state.State
	URL          string
	SwarmOptions swarm.SwarmOptions
}

type ErrSavingConfig struct {
	wrappedErr error
}

func (e ErrSavingConfig) Error() string {
	return fmt.Sprintf("Error saving config: %s", e.wrappedErr)
}

func NewHost(name, driverName string, hostOptions *HostOptions) (*Host, error) {
	authOptions := hostOptions.AuthOptions
	storePath := filepath.Join(utils.GetMachineDir(), name)
	driver, err := drivers.NewDriver(driverName, name, storePath, authOptions.CaCertPath, authOptions.PrivateKeyPath)
	if err != nil {
		return nil, err
	}
	return &Host{
		Name:          name,
		ConfigVersion: version.ConfigVersion,
		DriverName:    driverName,
		Driver:        driver,
		StorePath:     storePath,
		HostOptions:   hostOptions,
	}, nil
}

func LoadHost(name string, StorePath string) (*Host, error) {
	if _, err := os.Stat(StorePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("Host %q does not exist", name)
	}

	host := &Host{Name: name, StorePath: StorePath}
	if err := host.LoadConfig(); err != nil {
		return nil, err
	}

	return host, nil
}

func ValidateHostName(name string) bool {
	return validHostNamePattern.MatchString(name)
}

func (h *Host) Create(name string) error {
	// create the instance
	if err := h.Driver.Create(); err != nil {
		return err
	}

	// save to store
	if err := h.SaveConfig(); err != nil {
		return err
	}

	// TODO: Not really a fan of just checking "none" here.
	if h.Driver.DriverName() != "none" {
		if err := utils.WaitFor(drivers.MachineInState(h.Driver, state.Running)); err != nil {
			return err
		}

		if err := WaitForSSH(h); err != nil {
			return err
		}

		provisioner, err := provision.DetectProvisioner(h.Driver)
		if err != nil {
			return err
		}

		if err := provisioner.Provision(*h.HostOptions.SwarmOptions, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions); err != nil {
			return err
		}
	}

	return nil
}

func (h *Host) RunSSHCommand(command string) (string, error) {
	return drivers.RunSSHCommandFromDriver(h.Driver, command)
}

func (h *Host) CreateSSHClient() (ssh.Client, error) {
	addr, err := h.Driver.GetSSHHostname()
	if err != nil {
		return ssh.ExternalClient{}, err
	}

	port, err := h.Driver.GetSSHPort()
	if err != nil {
		return ssh.ExternalClient{}, err
	}

	auth := &ssh.Auth{
		Keys: []string{h.Driver.GetSSHKeyPath()},
	}

	return ssh.NewClient(h.Driver.GetSSHUsername(), addr, port, auth)
}

func (h *Host) CreateSSHShell() error {
	client, err := h.CreateSSHClient()
	if err != nil {
		return err
	}

	return client.Shell()
}

func (h *Host) runActionForState(action func() error, desiredState state.State) error {
	if drivers.MachineInState(h.Driver, desiredState)() {
		log.Debug("Machine already in state %s, returning", desiredState)
		return nil
	}

	if err := action(); err != nil {
		return err
	}

	if err := h.SaveConfig(); err != nil {
		return err
	}

	return utils.WaitFor(drivers.MachineInState(h.Driver, desiredState))
}

func (h *Host) Start() error {
	return h.runActionForState(h.Driver.Start, state.Running)
}

func (h *Host) Stop() error {
	return h.runActionForState(h.Driver.Stop, state.Stopped)
}

func (h *Host) Kill() error {
	return h.runActionForState(h.Driver.Kill, state.Stopped)
}

func (h *Host) Restart() error {
	if drivers.MachineInState(h.Driver, state.Running)() {
		if err := h.Stop(); err != nil {
			return err
		}

		if err := utils.WaitFor(drivers.MachineInState(h.Driver, state.Stopped)); err != nil {
			return err
		}
	}

	if err := h.Start(); err != nil {
		return err
	}

	if err := utils.WaitFor(drivers.MachineInState(h.Driver, state.Running)); err != nil {
		return err
	}

	if err := h.SaveConfig(); err != nil {
		return err
	}

	return nil
}

func (h *Host) Upgrade() error {
	machineState, err := h.Driver.GetState()
	if err != nil {
		return err
	}

	if machineState != state.Running {
		log.Fatal(errMachineMustBeRunningForUpgrade)
	}

	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	if err := provisioner.Package("docker", pkgaction.Upgrade); err != nil {
		return err
	}

	if err := provisioner.Service("docker", pkgaction.Restart); err != nil {
		return err
	}
	return nil
}

func (h *Host) Remove(force bool) error {
	if err := h.Driver.Remove(); err != nil {
		if !force {
			return err
		}
	}

	if err := h.SaveConfig(); err != nil {
		return err
	}

	return h.removeStorePath()
}

func (h *Host) removeStorePath() error {
	file, err := os.Stat(h.StorePath)
	if err != nil {
		return err
	}
	if !file.IsDir() {
		return fmt.Errorf("%q is not a directory", h.StorePath)
	}
	return os.RemoveAll(h.StorePath)
}

func (h *Host) GetURL() (string, error) {
	return h.Driver.GetURL()
}

// IsActive provides a single method for determining if a host is active based
// on both the url and if the host is stopped.
func (h *Host) IsActive() (bool, error) {
	currentState, err := h.Driver.GetState()

	if err != nil {
		log.Errorf("error getting state for host %s: %s", h.Name, err)
		return false, err
	}

	url, err := h.GetURL()

	if err != nil {
		if err == drivers.ErrHostIsNotRunning {
			url = ""
		} else {
			log.Errorf("error getting URL for host %s: %s", h.Name, err)
			return false, err
		}
	}

	dockerHost := os.Getenv("DOCKER_HOST")

	notStopped := currentState != state.Stopped
	correctURL := url == dockerHost

	isActive := notStopped && correctURL

	return isActive, nil
}

func (h *Host) LoadConfig() error {
	data, err := ioutil.ReadFile(filepath.Join(h.StorePath, "config.json"))
	if err != nil {
		return err
	}

	// Remember the machine name and store path so we don't have to pass it
	// through each struct in the migration.
	name := h.Name
	storePath := h.StorePath

	// If we end up performing a migration, we should save afterwards so we don't have to do it again on subsequent invocations.
	migrationPerformed := h.ConfigVersion != version.ConfigVersion

	migratedHost, err := MigrateHost(h, data)
	if err != nil {
		return fmt.Errorf("Error getting migrated host: %s", err)
	}

	*h = *migratedHost

	h.Name = name
	h.StorePath = storePath

	if migrationPerformed {
		if err := h.SaveConfig(); err != nil {
			return fmt.Errorf("Error saving config after migration was performed: %s", err)
		}
	}

	return nil
}

func (h *Host) ConfigureAuth() error {
	if err := h.LoadConfig(); err != nil {
		return err
	}

	provisioner, err := provision.DetectProvisioner(h.Driver)
	if err != nil {
		return err
	}

	// TODO: This is kind of a hack (or is it?  I'm not really sure until
	// we have more clearly defined outlook on what the responsibilities
	// and modularity of the provisioners should be).
	//
	// Call provision to re-provision the certs properly.
	if err := provisioner.Provision(swarm.SwarmOptions{}, *h.HostOptions.AuthOptions, *h.HostOptions.EngineOptions); err != nil {
		return err
	}

	return nil
}

func (h *Host) SaveConfig() error {
	data, err := json.Marshal(h)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(filepath.Join(h.StorePath, "config.json"), data, 0600); err != nil {
		return err
	}
	return nil
}

func (h *Host) PrintIP() error {
	if ip, err := h.Driver.GetIP(); err != nil {
		return err
	} else {
		fmt.Println(ip)
	}
	return nil
}

func WaitForSSH(h *Host) error {
	return drivers.WaitForSSH(h.Driver)
}

func attemptGetHostState(host Host, stateQueryChan chan<- HostListItem) {
	currentState, err := host.Driver.GetState()
	if err != nil {
		log.Errorf("error getting state for host %s: %s", host.Name, err)
	}

	url, err := host.GetURL()
	if err != nil {
		if err == drivers.ErrHostIsNotRunning {
			url = ""
		} else {
			log.Errorf("error getting URL for host %s: %s", host.Name, err)
		}
	}

	isActive, err := host.IsActive()
	if err != nil {
		log.Errorf("error determining if host is active for host %s: %s",
			host.Name, err)
	}

	stateQueryChan <- HostListItem{
		Name:         host.Name,
		Active:       isActive,
		DriverName:   host.Driver.DriverName(),
		State:        currentState,
		URL:          url,
		SwarmOptions: *host.HostOptions.SwarmOptions,
	}
}

func getHostState(host Host, hostListItemsChan chan<- HostListItem) {
	// This channel is used to communicate the properties we are querying
	// about the host in the case of a successful read.
	stateQueryChan := make(chan HostListItem)

	go attemptGetHostState(host, stateQueryChan)

	select {
	// If we get back useful information, great.  Forward it straight to
	// the original parent channel.
	case hli := <-stateQueryChan:
		hostListItemsChan <- hli

	// Otherwise, give up after a predetermined duration.
	case <-time.After(stateTimeoutDuration):
		hostListItemsChan <- HostListItem{
			Name:       host.Name,
			DriverName: host.Driver.DriverName(),
			State:      state.Timeout,
		}
	}
}

func GetHostListItems(hostList []*Host) []HostListItem {
	hostListItems := []HostListItem{}
	hostListItemsChan := make(chan HostListItem)

	for _, host := range hostList {
		go getHostState(*host, hostListItemsChan)
	}

	for range hostList {
		hostListItems = append(hostListItems, <-hostListItemsChan)
	}

	close(hostListItemsChan)
	return hostListItems
}

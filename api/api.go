package api

import (
	drivers "github.com/docker/machine/drivers"
	"github.com/docker/machine/store"
)

type Api struct {
	store *store.Store
}

func NewApi(rootPath string) *Api {
	store := store.NewStore(rootPath)
	return &Api{store: store}
}

func (a *Api) Create(name string, driverName string, flags drivers.DriverOptions) (*store.Host, error) {
	return a.store.Create(name, driverName, flags)
}

func (a *Api) Remove(name string, force bool) error {
	return a.store.Remove(name, force)
}

func (a *Api) List() ([]store.Host, error) {
	return a.store.List()
}

func (a *Api) Load(name string) (*store.Host, error) {
	return a.store.Load(name)
}

func (a *Api) GetActive() (*store.Host, error) {
	return a.store.GetActive()
}

func (a *Api) IsActive(host *store.Host) (bool, error) {
	return a.store.IsActive(host)
}

func (a *Api) SetActive(host *store.Host) error {
	return a.store.SetActive(host)
}

// Kills the host specified by name. If name is empty, the
// active host will be killed.
func (a *Api) Kill(name string) error {
	host, err := a.getHost(name)
	if err != nil {
		return err
	}

	return host.Driver.Restart()
}

// Restarts the host specified by name. If name is empty, the
// active host will be restarted.
func (a *Api) Restart(name string) error {
	host, err := a.getHost(name)
	if err != nil {
		return err
	}

	return host.Driver.Restart()
}

// Starts the host specified by name. If name is empty, the
// active host will be started.
func (a *Api) Start(name string) error {
	host, err := a.getHost(name)
	if err != nil {
		return err
	}

	return host.Driver.Start()
}

// Stops the host specified by name. If name is empty, the
// active host will be stopped.
func (a *Api) Stop(name string) error {
	host, err := a.getHost(name)
	if err != nil {
		return err
	}

	return host.Driver.Stop()
}

// Upgrades the host specified by name. If name is empty, the
// active host will be upgraded.
func (a *Api) Upgrade(name string) error {
	host, err := a.getHost(name)
	if err != nil {
		return err
	}

	return host.Driver.Upgrade()
}

func (a *Api) getHost(name string) (*store.Host, error) {
	if name == "" {
		host, err := a.GetActive()
		if err != nil {
			return nil, err
		}
		return host, nil
	}

	host, err := a.Load(name)
	if err != nil {
		return nil, err
	}
	return host, nil
}

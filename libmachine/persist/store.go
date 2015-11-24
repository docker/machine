package persist

import "github.com/docker/machine/libmachine/host"

type Store interface {
	// Exists returns whether a machine exists or not
	Exists(name string) (bool, error)

	// List returns a list of all hosts in the store
	List() ([]string, error)

	// Load loads a host by name
	Load(name string) (*host.Host, error)

	// Remove removes a machine from the store
	Remove(name string) error

	// Save persists a machine in the store
	Save(host *host.Host) error
}

func LoadHosts(s Store, hostNames []string) ([]*host.Host, error) {
	loadedHosts := []*host.Host{}

	for _, hostName := range hostNames {
		h, err := s.Load(hostName)
		if err != nil {
			// TODO: (nathanleclaire) Should these be bundled up
			// into one error instead of exiting?
			return nil, err
		}
		loadedHosts = append(loadedHosts, h)
	}

	return loadedHosts, nil
}

func LoadAllHosts(s Store) ([]*host.Host, error) {
	hostNames, err := s.List()
	if err != nil {
		return nil, err
	}

	return LoadHosts(s, hostNames)
}

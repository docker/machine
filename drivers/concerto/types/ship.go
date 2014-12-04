package types

type Ships struct {
	Ships []Ship
}

type Ship struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Fqdn       string `json:"fqdn"`
	Ip         string `json:"ipAddress"`
	State      string `json:"state"`
	Os         string `json:"os"`
	Plan       string `json:"plan"`
	Port       int    `json:"port"`
	Schema     string `json:"schema"`
	LocalPort  int
	Containers []Containers
	Touched    bool
}

func (ship *Ship) IsNil() bool {
	if ship.Id == "" {
		return true
	} else {
		return false
	}
}

type Port struct {
	PrivatePort int64
	PublicPort  int64
	Type        string
	IP          string
}

type Containers struct {
	ID         string `json:"Id"`
	Image      string
	Command    string
	Created    int64
	Status     string
	Ports      []Port
	SizeRw     int64
	SizeRootFs int64
	Names      []string
}

type Clouds struct {
	Plans []Plan
}

type Plan struct {
	Id        string `json:"id"`
	Provider  string `json:"cloud_provider"`
	Continent string `json:"continent"`
	Region    string `json:"region"`
	Plan      string `json:"plan"`
}

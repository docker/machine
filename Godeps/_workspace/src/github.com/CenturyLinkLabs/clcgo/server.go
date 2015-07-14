package clcgo

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	serverCreationURL       = apiRoot + "/servers/%s"
	serverURL               = serverCreationURL + "/%s"
	publicIPAddressURL      = serverURL + "/publicIPAddresses"
	serverActiveStatus      = "active"
	serverStartedPowerState = "started"
	serverPausedPowerState  = "paused"
)

// A Server can be used to either fetch an existing Server or provision and new
// one. To fetch, you must supply an ID value. For creation, there are numerous
// required values. The API documentation should be consulted.
//
// The SourceServerID is a required field that allows multiple values which are
// documented in the API. One of the allowed values is a Template ID, which can
// be retrieved with the DataCenterCapabilities resource.
//
// To make your server a member of a specific network, you can set the
// DeployableNetwork field. This is optional. The Server will otherwise be a
// member of the default network. DeployableNetworks exist per account and
// DataCenter and can be retrieved via the DataCenterCapabilities resource. If
// you know the NetworkID, you can supply it instead.
//
// A Password field can be set for a Server you are saving, but fetching the
// username and password for an existing Server can only be done via the
// Credentials resource.
type Server struct {
	uuidURI           string            `json:"-"`
	ID                string            `json:"id"`
	Name              string            `json:"name"`
	GroupID           string            `json:"groupId"`
	Status            string            `json:"status"`
	SourceServerID    string            `json:"sourceServerId"` // TODO: nonexistent in get, extract to creation params?
	CPU               int               `json:"cpu"`
	MemoryGB          int               `json:"memoryGB"` // TODO: memoryMB in get, extract to creation params?
	Type              string            `json:"type"`
	DeployableNetwork DeployableNetwork `json:"-"`
	NetworkID         string            `json:"networkId"`
	Password          string            `json:"password"`
	Details           struct {
		PowerState  string `json:"powerState"`
		IPAddresses []struct {
			Public   string `json:"public"`
			Internal string `json:"internal"`
		} `json:"ipAddresses"`
	} `json:"details"`
}

// Credentials can be used to fetch the username and password for a Server. You
// must supply the associated Server.
//
// This uses an undocumented API endpoint and could be changed or removed.
type Credentials struct {
	Server   Server `json:"-"`
	Username string `json:"userName"`
	Password string `json:"password"`
}

type serverCreationResponse struct {
	Links []Link `json:"links"`
}

// A PublicIPAddress can be created and associated with an existing,
// provisioned Server. You must supply the associated Server object.
//
// You must supply a slice of Port objects that will make the specified ports
// accessible at the address.
type PublicIPAddress struct {
	Server            Server
	Ports             []Port `json:"ports"`
	InternalIPAddress string `json:"internalIPAddress"`
}

// A Port object specifies a network port that should be made available on a
// PublicIPAddress. It can only be used in conjunction with the PublicIPAddress
// resource.
type Port struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
}

// IsActive will, unsurprisingly, tell you if the Server is both active and not
// paused.
func (s Server) IsActive() bool {
	return s.Status == serverActiveStatus && s.Details.PowerState == serverStartedPowerState
}

// IsPaused will tell you if the Server is paused or not.
func (s Server) IsPaused() bool {
	return s.Details.PowerState == serverPausedPowerState
}

func (s Server) URL(a string) (string, error) {
	if s.ID == "" && s.uuidURI == "" {
		return "", errors.New("an ID field is required to get a server")
	} else if s.uuidURI != "" {
		return apiDomain + s.uuidURI, nil
	}

	return fmt.Sprintf(serverURL, a, s.ID), nil
}

func (s *Server) RequestForSave(a string) (request, error) {
	url := fmt.Sprintf(serverCreationURL, a)
	s.NetworkID = s.DeployableNetwork.NetworkID
	return request{URL: url, Parameters: *s}, nil
}

func (s *Server) StatusFromCreateResponse(r []byte) (Status, error) {
	scr := serverCreationResponse{}
	err := json.Unmarshal(r, &scr)
	if err != nil {
		return Status{}, err
	}

	sl, err := typeFromLinks("status", scr.Links)
	if err != nil {
		return Status{}, errors.New("the creation response has no status link")
	}

	il, err := typeFromLinks("self", scr.Links)
	if err != nil {
		return Status{}, errors.New("the creation response has no self link")
	}

	s.uuidURI = il.HRef

	return Status{URI: sl.HRef}, nil
}

func (c Credentials) URL(a string) (string, error) {
	if c.Server.ID == "" {
		return "", errors.New("a Server with an ID is required to fetch credentials")
	}

	url := fmt.Sprintf("%s/servers/%s/%s/credentials", apiRoot, a, c.Server.ID)
	return url, nil
}

func (i PublicIPAddress) RequestForSave(a string) (request, error) {
	if i.Server.ID == "" {
		return request{}, errors.New("a Server with an ID is required to add a Public IP Address")
	}

	url := fmt.Sprintf(publicIPAddressURL, a, i.Server.ID)
	return request{URL: url, Parameters: i}, nil
}

func (i PublicIPAddress) StatusFromCreateResponse(r []byte) (Status, error) {
	l := Link{}
	err := json.Unmarshal(r, &l)
	if err != nil {
		return Status{}, err
	}

	return Status{URI: l.HRef}, nil
}

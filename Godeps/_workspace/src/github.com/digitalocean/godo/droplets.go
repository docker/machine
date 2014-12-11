package godo

import "fmt"

const dropletBasePath = "v2/droplets"

// DropletsService is an interface for interfacing with the droplet
// endpoints of the Digital Ocean API
// See: https://developers.digitalocean.com/#droplets
type DropletsService interface {
	List(*ListOptions) ([]Droplet, *Response, error)
	Get(int) (*DropletRoot, *Response, error)
	Create(*DropletCreateRequest) (*DropletRoot, *Response, error)
	Delete(int) (*Response, error)
	dropletActionStatus(string) (string, error)
}

// DropletsServiceOp handles communication with the droplet related methods of the
// DigitalOcean API.
type DropletsServiceOp struct {
	client *Client
}

// Droplet represents a DigitalOcean Droplet
type Droplet struct {
	ID          int       `json:"id,float64,omitempty"`
	Name        string    `json:"name,omitempty"`
	Memory      int       `json:"memory,omitempty"`
	Vcpus       int       `json:"vcpus,omitempty"`
	Disk        int       `json:"disk,omitempty"`
	Region      *Region   `json:"region,omitempty"`
	Image       *Image    `json:"image,omitempty"`
	Size        *Size     `json:"size,omitempty"`
	BackupIDs   []int     `json:"backup_ids,omitempty"`
	SnapshotIDs []int     `json:"snapshot_ids,omitempty"`
	Locked      bool      `json:"locked,bool,omitempty"`
	Status      string    `json:"status,omitempty"`
	Networks    *Networks `json:"networks,omitempty"`
	ActionIDs   []int     `json:"action_ids,omitempty"`
}

// Convert Droplet to a string
func (d Droplet) String() string {
	return Stringify(d)
}

// DropletRoot represents a Droplet root
type DropletRoot struct {
	Droplet *Droplet `json:"droplet"`
	Links   *Links   `json:"links,omitempty"`
}

type dropletsRoot struct {
	Droplets []Droplet `json:"droplets"`
	Links    *Links    `json:"links"`
}

// DropletCreateRequest represents a request to create a droplet.
type DropletCreateRequest struct {
	Name              string        `json:"name"`
	Region            string        `json:"region"`
	Size              string        `json:"size"`
	Image             string        `json:"image"`
	SSHKeys           []interface{} `json:"ssh_keys"`
	Backups           bool          `json:"backups"`
	IPv6              bool          `json:"ipv6"`
	PrivateNetworking bool          `json:"private_networking"`
	UserData          string        `json:"user_data"`
}

func (d DropletCreateRequest) String() string {
	return Stringify(d)
}

// Networks represents the droplet's networks
type Networks struct {
	V4 []Network `json:"v4,omitempty"`
	V6 []Network `json:"v6,omitempty"`
}

// Network represents a DigitalOcean Network
type Network struct {
	IPAddress string `json:"ip_address,omitempty"`
	Netmask   string `json:"netmask,omitempty"`
	Gateway   string `json:"gateway,omitempty"`
	Type      string `json:"type,omitempty"`
}

func (n Network) String() string {
	return Stringify(n)
}

// List all droplets
func (s *DropletsServiceOp) List(opt *ListOptions) ([]Droplet, *Response, error) {
	path := dropletBasePath
	path, err := addOptions(path, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(dropletsRoot)
	resp, err := s.client.Do(req, root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}

	return root.Droplets, resp, err
}

// Get individual droplet
func (s *DropletsServiceOp) Get(dropletID int) (*DropletRoot, *Response, error) {
	path := fmt.Sprintf("%s/%d", dropletBasePath, dropletID)

	req, err := s.client.NewRequest("GET", path, nil)
	if err != nil {
		return nil, nil, err
	}

	root := new(DropletRoot)
	resp, err := s.client.Do(req, root)
	if err != nil {
		return nil, resp, err
	}

	return root, resp, err
}

// Create droplet
func (s *DropletsServiceOp) Create(createRequest *DropletCreateRequest) (*DropletRoot, *Response, error) {
	path := dropletBasePath

	req, err := s.client.NewRequest("POST", path, createRequest)
	if err != nil {
		return nil, nil, err
	}

	root := new(DropletRoot)
	resp, err := s.client.Do(req, root)
	if err != nil {
		return nil, resp, err
	}
	if l := root.Links; l != nil {
		resp.Links = l
	}

	return root, resp, err
}

// Delete droplet
func (s *DropletsServiceOp) Delete(dropletID int) (*Response, error) {
	path := fmt.Sprintf("%s/%d", dropletBasePath, dropletID)

	req, err := s.client.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)

	return resp, err
}

func (s *DropletsServiceOp) dropletActionStatus(uri string) (string, error) {
	action, _, err := s.client.DropletActions.GetByURI(uri)

	if err != nil {
		return "", err
	}

	return action.Status, nil
}

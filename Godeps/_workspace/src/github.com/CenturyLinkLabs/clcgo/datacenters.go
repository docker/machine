package clcgo

import (
	"errors"
	"fmt"
)

const (
	dataCentersURL            = apiRoot + "/datacenters/%s"
	dataCenterGroupURL        = dataCentersURL + "/%s?groupLinks=true"
	dataCenterCapabilitiesURL = dataCentersURL + "/%s/deploymentCapabilities"
)

// The DataCenters resource can retrieve a list of available DataCenters.
type DataCenters []DataCenter

// A DataCenter resource can either be returned by the DataCenters resource, or
// built manually. It should be used in conjunction with the
// DataCenterCapabilities resource to request information about it.
//
// You must supply the ID if you are building this object manually.
type DataCenter struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// DataCenterGroup can be used to find associated groups for a datacenter and
// your account. The linked hardware group can be used with the Group object to
// query for your account's groups within that datacenter.
//
// You must supply the associated DataCenter object.
type DataCenterGroup struct {
	DataCenter DataCenter `json:"-"`
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Links      []Link     `json:"links"`
}

// DataCenterCapabilities gets more information about a specific DataCenter.
// You must supply the associated DataCenter object.
type DataCenterCapabilities struct {
	DataCenter         DataCenter          `json:"-"`
	DeployableNetworks []DeployableNetwork `json:"deployableNetworks"`
	Templates          []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"templates"`
}

// A DeployableNetwork describes a private network that is scoped to an Account
// and DataCenter. It can be fetched via the DataCenterCapabilities and can
// optionally be used to put a Server in a specific network.
type DeployableNetwork struct {
	Name      string `json:"name"`
	NetworkID string `json:"networkId"`
	Type      string `json:"type"`
	AccountID string `json:"accountID"`
}

func (d DataCenters) URL(a string) (string, error) {
	return fmt.Sprintf(dataCentersURL, a), nil
}

func (d DataCenterCapabilities) URL(a string) (string, error) {
	if d.DataCenter.ID == "" {
		return "", errors.New("need a DataCenter with an ID")
	}

	return fmt.Sprintf(dataCenterCapabilitiesURL, a, d.DataCenter.ID), nil
}

func (d DataCenterGroup) URL(a string) (string, error) {
	if d.DataCenter.ID == "" {
		return "", errors.New("need a DataCenter with an ID")
	}

	return fmt.Sprintf(dataCenterGroupURL, a, d.DataCenter.ID), nil
}

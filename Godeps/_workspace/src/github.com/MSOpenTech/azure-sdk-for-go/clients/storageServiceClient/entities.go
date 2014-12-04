package storageServiceClient

import (
	"encoding/xml"
)

type StorageServiceList struct {
	XMLName         xml.Name         `xml:"StorageServices"`
	Xmlns           string           `xml:"xmlns,attr"`
	StorageServices []StorageService `xml:"StorageService"`
}

type StorageService struct {
	Url                      string
	ServiceName              string
	StorageServiceProperties StorageServiceProperties
}

type StorageServiceProperties struct {
	Description           string
	Location              string
	Label                 string
	Status                string
	Endpoints             []string `xml:"Endpoints>Endpoint"`
	GeoReplicationEnabled string
	GeoPrimaryRegion      string
}

type StorageServiceDeployment struct {
	XMLName               xml.Name `xml:"CreateStorageServiceInput"`
	Xmlns                 string   `xml:"xmlns,attr"`
	ServiceName           string
	Description           string
	Label                 string
	AffinityGroup         string `xml:",omitempty"`
	Location              string `xml:",omitempty"`
	GeoReplicationEnabled bool
	ExtendedProperties    ExtendedPropertyList
	SecondaryReadEnabled  bool
}

type ExtendedPropertyList struct {
	ExtendedProperty []ExtendedProperty
}

type ExtendedProperty struct {
	Name  string
	Value string
}

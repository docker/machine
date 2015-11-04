package locationClient

import (
	"encoding/xml"
)

type LocationList struct {
	XMLName   xml.Name   `xml:"Locations"`
	Xmlns     string     `xml:"xmlns,attr"`
	Locations []Location `xml:"Location"`
}

type Location struct {
	Name              string
	DisplayName       string
	AvailableServices []string `xml:"AvailableServices>AvailableService"`
}

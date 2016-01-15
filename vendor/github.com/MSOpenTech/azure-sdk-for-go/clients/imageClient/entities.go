package imageClient

import (
	"encoding/xml"
)

type ImageList struct {
	XMLName  xml.Name  `xml:"Images"`
	Xmlns    string    `xml:"xmlns,attr"`
	OSImages []OSImage `xml:"OSImage"`
}

type OSImage struct {
	Category        string
	Label           string
	LogicalSizeInGB string
	Name            string
	OS              string
	Eula            string
	Description     string
	Location        string
}

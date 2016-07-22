package cloud

import (
	"time"
)

type IpAddress struct {
	Status       string    `json:"status,omitempty"`
	Version      string    `json:"version,omitempty"`
	ZoneId       string    `json:"zoneId,omitempty"`
	CreatedAt    time.Time `json:"createdAt,omitempty"`
	Address      string    `json:"address,omitempty"`
	AddressId    string    `json:"addressId,omitempty"`
	Type_        string    `json:"type,omitempty"`
	DataCenterId string    `json:"dataCenterId,omitempty"`
	ModifiedAt   time.Time `json:"modifiedAt,omitempty"`
	ServerId     string    `json:"serverId,omitempty"`
}

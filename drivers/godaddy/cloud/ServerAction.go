package cloud

import (
	"time"
)

type ServerAction struct {
	Status         string     `json:"status,omitempty"`
	ServerActionId string     `json:"serverActionId,omitempty"`
	CreatedAt      time.Time  `json:"createdAt,omitempty"`
	ModifiedAt     time.Time  `json:"modifiedAt,omitempty"`
	Type_          string     `json:"type,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	ServerId       string     `json:"serverId,omitempty"`
}

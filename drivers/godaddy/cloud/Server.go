package cloud

import (
	"time"
)

type Server struct {
	PrivateIp        string    `json:"privateIp,omitempty"`
	ImageId          string    `json:"imageId,omitempty"`
	BackupsEnabled   bool      `json:"backupsEnabled,omitempty"`
	Description      string    `json:"description,omitempty"`
	SpecId           string    `json:"specId,omitempty"`
	PublicIp         string    `json:"publicIp,omitempty"`
	DataCenterId     string    `json:"dataCenterId,omitempty"`
	ModifiedAt       time.Time `json:"modifiedAt,omitempty"`
	Status           string    `json:"status,omitempty"`
	BackupScheduleId string    `json:"backupScheduleId,omitempty"`
	CreatedAt        time.Time `json:"createdAt,omitempty"`
	ZoneId           string    `json:"zoneId,omitempty"`
	Hostname         string    `json:"hostname,omitempty"`
	Username         string    `json:"username,omitempty"`
	SshKeyId         string    `json:"sshKeyId,omitempty"`
	TaskState        string    `json:"taskState,omitempty"`
	ServerId         string    `json:"serverId,omitempty"`
}

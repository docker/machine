package cloud

type ServerCreate struct {
	BackupsEnabled bool     `json:"backupsEnabled,omitempty"`
	SshKeyId       string   `json:"sshKeyId,omitempty"`
	Volumes        []string `json:"volumes,omitempty"`
	DataCenterId   string   `json:"dataCenterId,omitempty"`
	Addresses      []string `json:"addresses,omitempty"`
	Spec           string   `json:"spec,omitempty"`
	Description    string   `json:"description,omitempty"`
	ZoneId         string   `json:"zoneId,omitempty"`
	Password       string   `json:"password,omitempty"`
	Hostname       string   `json:"hostname,omitempty"`
	BootScript     string   `json:"bootScript,omitempty"`
	Username       string   `json:"username,omitempty"`
	Discount       string   `json:"discount,omitempty"`
	Image          string   `json:"image,omitempty"`
}

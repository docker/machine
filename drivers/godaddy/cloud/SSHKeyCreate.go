package cloud

type SSHKeyCreate struct {
	Key  string `json:"key,omitempty"`
	Name string `json:"name,omitempty"`
}

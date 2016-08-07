package cloud

import (
	"time"
)

type SSHKey struct {
	Fingerprint string    `json:"fingerprint,omitempty"`
	Name        string    `json:"name,omitempty"`
	SshKeyId    string    `json:"sshKeyId,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty"`
	ModifiedAt  time.Time `json:"modifiedAt,omitempty"`
}

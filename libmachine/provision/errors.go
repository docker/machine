package provision

import (
	"errors"
)

var (
	ErrDetectionFailed  = errors.New("OS type not recognized")
	ErrSSHCommandFailed = errors.New("SSH command failure")
	ErrNotImplemented   = errors.New("Runtime not implemented")
)

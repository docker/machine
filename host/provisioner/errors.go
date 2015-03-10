package provisioner

import (
	"errors"
)

var (
	ErrDetectionFailed  = errors.New("Runtime detection failed")
	ErrSSHCommandFailed = errors.New("SSH command failure")
	ErrNotImplemented   = errors.New("Runtime not implemented")
)

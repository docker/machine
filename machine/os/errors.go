package os

import (
	"errors"
)

var (
	ErrDetectionFailed  = errors.New("Runtime detection failed")
	ErrSSHCommandFailed = errors.New("SSH command failure")
)

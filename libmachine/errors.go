package libmachine

import (
	"errors"
)

var (
	ErrHostDoesNotExist    = errors.New("Host does not exist")
	ErrInvalidHostname     = errors.New("Invalid hostname specified")
	ErrUnknownProviderType = errors.New("Unknown hypervisor type")
)

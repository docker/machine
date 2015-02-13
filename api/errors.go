package api

import (
	"errors"
)

var (
	ErrMachineExists       = errors.New("machine exists")
	ErrMachineDoesNotExist = errors.New("machine does not exist")
	ErrInvalidHostname     = errors.New("invalid hostname specified")
)

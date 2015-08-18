package mcnerror

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidHostname     = errors.New("Invalid hostname specified")
	ErrUnknownProviderType = errors.New("Unknown hypervisor type")
)

type ErrHostDoesNotExist struct {
	Name string
}

func (e ErrHostDoesNotExist) Error() string {
	return fmt.Sprintf("Host does not exist: %q", e.Name)
}

type ErrHostAlreadyExists struct {
	Name string
}

func (e ErrHostAlreadyExists) Error() string {
	return fmt.Sprintf("Host already exists: %q", e.Name)
}

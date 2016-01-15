package mcnerror

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidHostname = errors.New("Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -")
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

type ErrDuringPreCreate struct {
	Cause error
}

func (e ErrDuringPreCreate) Error() string {
	return fmt.Sprintf("Error with pre-create check: %q", e.Cause)
}

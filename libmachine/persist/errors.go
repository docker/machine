package persist

import "fmt"

type SaveError struct {
	hostName string
	reason   error
}

func NewSaveError(hostName string, reason error) error {
	err := SaveError{
		hostName: hostName,
		reason:   reason,
	}
	return &err
}

func (err *SaveError) Error() string {
	return fmt.Sprintf("Error attempting to save host %q to store: %s", err.hostName, err.reason)
}

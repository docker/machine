package clcgo

import (
	"encoding/json"
	"errors"
	"fmt"
)

// An OperationType should be used in conjunction with a ServerOperation to
// instruct a server instance to perform one of several operations.
type OperationType string

// More information on precisely what each operation does can be found in the
// CenturyLink Cloud V2 API documentation.
const (
	PauseServer    OperationType = "pause"
	ShutDownServer OperationType = "shutDown"
	RebootServer   OperationType = "reboot"
	ResetServer    OperationType = "reset"
	PowerOnServer  OperationType = "powerOn"
	PowerOffServer OperationType = "powerOff"
)

// A ServerOperation is used to perform many tasks around
// starting/stopping/pausing server instances.
//
// You are required to pass both a Server reference and an OperationType, and
// you will be notified of the operation's progress via the Status that
// is returned when the ServerOperation is passed to SaveEntity.
type ServerOperation struct {
	OperationType OperationType
	Server        Server
}

type operationResponse struct {
	Links []Link `json:"links"`
}

const (
	operationsRoot = apiRoot + "/operations"
	operationURL   = operationsRoot + "/%s/servers/%s"
)

func (p ServerOperation) RequestForSave(a string) (request, error) {
	if p.Server.ID == "" || p.OperationType == "" {
		return request{}, errors.New("a ServerOperation requires a Server and OperationType")
	}

	r := request{
		URL:        fmt.Sprintf(operationURL, a, p.OperationType),
		Parameters: []string{p.Server.ID},
	}

	return r, nil
}

func (p ServerOperation) StatusFromCreateResponse(r []byte) (Status, error) {
	ors := []operationResponse{}
	err := json.Unmarshal(r, &ors)
	if err != nil {
		return Status{}, err
	}

	// This API is capable of operating on multiple servers in one call, but
	// allowing for single or multiple entities everywhere is going to require
	// more reshuffling than we need to do right now. We enforce a single server
	// on the operation, and error if the API returns more than one operation. It
	// shouldn't based on us only submitting one server a time, but just in case.
	// I'd rather error than ignore or panic.
	if len(ors) != 1 {
		return Status{}, errors.New("expected a single operation response from the API!")
	}
	or := ors[0]

	sl, err := typeFromLinks("status", or.Links)
	if err != nil {
		return Status{}, errors.New("the operation response has no status link")
	}

	return Status{URI: sl.HRef}, nil
}

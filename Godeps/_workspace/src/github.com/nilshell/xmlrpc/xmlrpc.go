package xmlrpc

import (
	"fmt"
)

// Struct presents hash type used in xmlprc requests and responses.
type Struct map[string]interface{}

// Base64 represents base64 data
type Base64 string

// Params represents a list of parameters to a method.
type Params struct {
	Params []interface{}
}

// xmlrpcError represents errors returned on xmlrpc request.
type xmlrpcError struct {
	code    string
	message string
}

// Error() method implements Error interface
func (e *xmlrpcError) Error() string {
	return fmt.Sprintf("Error: \"%s\" Code: %s", e.message, e.code)
}

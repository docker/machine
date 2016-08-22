package driverutil

import "strings"

// SplitPortProto splits a string in the format port/protocol, defaulting
// protocol to "tcp" if not provided.
func SplitPortProto(raw string) (port string, protocol string, err error) {
	parts := strings.Split(raw, "/")
	if len(parts) == 1 {
		protocol = "tcp"
	} else {
		protocol = parts[1]
	}
	port = parts[0]
	return port, protocol, nil
}

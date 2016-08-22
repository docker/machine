package driverutil

import (
	"fmt"
	"strconv"
	"strings"
)

// SplitPortProto splits a string in the format port/protocol, defaulting
// protocol to "tcp" if not provided.
func SplitPortProto(raw string) (port int, protocol string, err error) {
	parts := strings.Split(raw, "/")
	if len(parts) == 1 {
		protocol = "tcp"
	} else {
		protocol = parts[1]
	}
	port, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, "", fmt.Errorf("invalid port number %s: %s", parts[0], err)
	}
	return port, protocol, nil
}

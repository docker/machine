package driverutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitPortProtocol(t *testing.T) {
	tests := []struct {
		raw           string
		expectedPort  string
		expectedProto string
		expectedErr   bool
	}{
		{"8080/tcp", "8080", "tcp", false},
		{"90/udp", "90", "udp", false},
		{"80", "80", "tcp", false},
	}

	for _, tc := range tests {
		port, proto, err := SplitPortProto(tc.raw)
		assert.Equal(t, tc.expectedPort, port)
		assert.Equal(t, tc.expectedProto, proto)
		if tc.expectedErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}
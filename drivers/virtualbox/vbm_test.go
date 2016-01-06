package virtualbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidCheckVBoxManageVersion(t *testing.T) {
	var tests = []struct {
		version string
	}{
		{"5.1"},
		{"5.0.8r103449"},
		{"5.0"},
		{"4.10"},
		{"4.3.1"},
	}

	for _, test := range tests {
		err := checkVBoxManageVersion(test.version)

		assert.NoError(t, err)
	}
}

func TestInvalidCheckVBoxManageVersion(t *testing.T) {
	var tests = []struct {
		version       string
		expectedError string
	}{
		{"3.9", `We support Virtualbox starting with version 5. Your VirtualBox install is "3.9". Please upgrade at https://www.virtualbox.org`},
		{"4.0", `We support Virtualbox starting with version 5. Your VirtualBox install is "4.0". Please upgrade at https://www.virtualbox.org`},
		{"4.1.1", `We support Virtualbox starting with version 5. Your VirtualBox install is "4.1.1". Please upgrade at https://www.virtualbox.org`},
		{"4.2.36-104064", `We support Virtualbox starting with version 5. Your VirtualBox install is "4.2.36-104064". Please upgrade at https://www.virtualbox.org`},
		{"X.Y", `We support Virtualbox starting with version 5. Your VirtualBox install is "X.Y". Please upgrade at https://www.virtualbox.org`},
		{"", `We support Virtualbox starting with version 5. Your VirtualBox install is "". Please upgrade at https://www.virtualbox.org`},
	}

	for _, test := range tests {
		err := checkVBoxManageVersion(test.version)

		assert.EqualError(t, err, test.expectedError)
	}
}

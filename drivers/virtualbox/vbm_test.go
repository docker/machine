package virtualbox

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckVBoxManageVersionValid(t *testing.T) {
	var tests = []struct {
		version string
	}{
		{"5.0.8r103449"},
		{"5.0"},
		{"5.1"},
		{"4.1"},
		{"4.2.0"},
		{"4.3.1"},
	}

	for _, test := range tests {
		err := checkVBoxManageVersion(test.version)

		assert.NoError(t, err)
	}
}

func TestCheckVBoxManageVersionInvalid(t *testing.T) {
	var tests = []struct {
		version       string
		expectedError string
	}{
		{"3.9", `We support Virtualbox starting with version 5. Your VirtualBox install is "3.9". Please upgrade at https://www.virtualbox.org`},
		{"", `We support Virtualbox starting with version 5. Your VirtualBox install is "". Please upgrade at https://www.virtualbox.org`},
	}

	for _, test := range tests {
		err := checkVBoxManageVersion(test.version)

		assert.EqualError(t, err, test.expectedError)
	}
}

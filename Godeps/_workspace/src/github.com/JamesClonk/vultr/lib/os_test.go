package lib

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_OS_GetOS_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	os, err := client.GetOS()
	assert.Nil(t, os)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_OS_GetOS_NoOS(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	os, err := client.GetOS()
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, os)
}

func Test_OS_GetOS_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{
"127":{"OSID":127,"name":"CentOS 6 x64","arch":"x64","family":"centos","windows":false},
"179":{"OSID":179,"name":"CoreOS Stable","arch":"x64","family":"coreos","windows":false},
"124":{"OSID":124,"name":"Windows 2012 R2 x64","arch":"x64","family":"windows","windows":true}}`)
	defer server.Close()

	os, err := client.GetOS()
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, os) {
		assert.Equal(t, 3, len(os))
		// OS could be in random order
		for _, os := range os {
			switch os.ID {
			case 127:
				assert.Equal(t, "CentOS 6 x64", os.Name)
				assert.Equal(t, "x64", os.Arch)
				assert.Equal(t, "centos", os.Family)
			case 179:
				assert.Equal(t, "coreos", os.Family)
				assert.Equal(t, "CoreOS Stable", os.Name)
				assert.Equal(t, false, os.Windows)
			case 124:
				assert.Equal(t, "windows", os.Family)
				assert.Equal(t, "Windows 2012 R2 x64", os.Name)
				assert.Equal(t, true, os.Windows)
			default:
				t.Error("Unknown OSID")
			}
		}
	}
}

package lib

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_SSHKeys_GetSSHKeys_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	keys, err := client.GetSSHKeys()
	assert.Nil(t, keys)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_SSHKeys_GetSSHKeys_NoKeys(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	keys, err := client.GetSSHKeys()
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, keys)
}

func Test_SSHKeys_GetSSHKeys_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{
"one":{"SSHKEYID":"1","name":"alpha","ssh_key":"aaaa","date_created":null},
"two":{"SSHKEYID":"2","name":"beta","ssh_key":"bbbb","date_created":"2014-12-31 13:34:56"},
"three":{"SSHKEYID":"3","name":"charlie","ssh_key":"cccc"}}`)
	defer server.Close()

	keys, err := client.GetSSHKeys()
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, keys) {
		assert.Equal(t, 3, len(keys))
		// keys could be in random order
		for _, key := range keys {
			switch key.ID {
			case "1":
				assert.Equal(t, "alpha", key.Name)
				assert.Equal(t, "", key.Created)
			case "2":
				assert.Equal(t, "beta", key.Name)
				assert.Equal(t, "2014-12-31 13:34:56", key.Created)
			case "3":
				assert.Equal(t, "cccc", key.Key)
			default:
				t.Error("Unknown SSHKEYID")
			}
		}
	}
}

func Test_SSHKeys_CreateSSHKey_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	key, err := client.CreateSSHKey("delta", "ddddd")
	assert.Equal(t, SSHKey{}, key)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_SSHKeys_CreateSSHKey_NoKey(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	key, err := client.CreateSSHKey("delta", "ddddd")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "", key.ID)
}

func Test_SSHKeys_CreateSSHKey_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{"SSHKEYID":"a1b2c3d4"}`)
	defer server.Close()

	key, err := client.CreateSSHKey("delta", "ddddd")
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, key) {
		assert.Equal(t, "a1b2c3d4", key.ID)
		assert.Equal(t, "delta", key.Name)
		assert.Equal(t, "ddddd", key.Key)
	}
}

func Test_SSHKeys_UpdateSSHKey_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.UpdateSSHKey(SSHKey{"o1", "omega", "oooo", "2012-12-12 12:12:12"})
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_SSHKeys_UpdateSSHKey_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	key := SSHKey{"o1", "omega", "oooo", "2012-12-12 12:12:12"}
	if err := client.UpdateSSHKey(key); err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, key) {
		assert.Equal(t, "o1", key.ID)
		assert.Equal(t, "omega", key.Name)
		assert.Equal(t, "oooo", key.Key)
		assert.Equal(t, "2012-12-12 12:12:12", key.Created)
	}
}

func Test_SSHKeys_DeleteSSHKey_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	err := client.DeleteSSHKey("id-1")
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_SSHKeys_DeleteSSHKey_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{no-response?!}`)
	defer server.Close()

	assert.Nil(t, client.DeleteSSHKey("id-1"))
}

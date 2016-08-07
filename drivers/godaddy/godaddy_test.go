package godaddy

import (
	"testing"

	"github.com/docker/machine/drivers/godaddy/cloud"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type clientMock struct {
	mock.Mock
}

func (m clientMock) SSHKeys() *cloud.SshKeysApi {
	return nil
}

type mockServersAPI struct {
	*cloud.ServersApi
	mock.Mock
}

func (m *clientMock) Servers() cloud.ServersClient {
	args := m.Called()
	return args.Get(0).(cloud.ServersClient)
}

func (m *mockServersAPI) AddServer(server cloud.ServerCreate) (cloud.Server, error) {
	args := m.Called(server)
	return cloud.Server{
		ServerId: args.String(0),
		Status:   "NEW",
	}, nil
}

func (m *mockServersAPI) StartServer(serverID string) (cloud.ServerAction, error) {
	args := m.Called(serverID)
	return args.Get(0).(cloud.ServerAction), args.Error(1)
}

func (m *mockServersAPI) StopServer(serverID string) (cloud.ServerAction, error) {
	args := m.Called(serverID)
	return args.Get(0).(cloud.ServerAction), args.Error(1)
}

func (m *mockServersAPI) GetServerById(serverID string) (cloud.Server, error) {
	args := m.Called(serverID)
	return cloud.Server{
		ServerId: serverID,
		PublicIp: "0.0.0.0",
		Status:   args.String(0),
	}, nil
}

func (m *mockServersAPI) DestroyServer(serverID string) (cloud.ServerAction, error) {
	args := m.Called(serverID)
	return args.Get(0).(cloud.ServerAction), args.Error(1)
}

func newTestDriver() (*Driver, *clientMock) {
	driver := NewDriver("hostName", "someStorePath")
	driver.ServerID = "0h18t27z"
	driver.SSHKey = "~/some/key"
	client := new(clientMock)
	driver.client = func() cloud.ClientWrapper {
		return client
	}
	driver.createSSHKey = func() error {
		driver.SSHKey = driver.GetSSHKeyPath()
		return nil
	}
	return driver, client
}

func TestSetConfigFromFlags(t *testing.T) {
	driver := NewDriver("default", "path")

	checkFlags := &drivers.CheckDriverOptions{
		FlagsValues: map[string]interface{}{
			"godaddy-api-key":    "my-api-key",
			"godaddy-ssh-key-id": "ssh-key",
		},
		CreateFlags: driver.GetCreateFlags(),
	}

	err := driver.SetConfigFromFlags(checkFlags)

	assert.NoError(t, err)
	assert.Empty(t, checkFlags.InvalidFlags)
}

func TestCreate(t *testing.T) {
	driver, client := newTestDriver()
	servers := new(mockServersAPI)
	newid := "newid123"
	client.On("Servers").Return(servers)
	servers.On("AddServer", mock.AnythingOfType("cloud.ServerCreate")).Return(newid)
	servers.On("GetServerById", newid).Return("RUNNING")

	err := driver.Create()
	assert.NoError(t, err)
}

func TestStart(t *testing.T) {
	driver, client := newTestDriver()
	servers := new(mockServersAPI)
	client.On("Servers").Return(servers)
	servers.On("StartServer", driver.ServerID).Return(cloud.ServerAction{}, nil)
	servers.On("GetServerById", driver.ServerID).Return("RUNNING")

	err := driver.Start()
	servers.AssertCalled(t, "StartServer", driver.ServerID)
	assert.NoError(t, err)
}

func TestKill(t *testing.T) {
	driver, client := newTestDriver()
	servers := new(mockServersAPI)
	client.On("Servers").Return(servers)
	servers.On("StopServer", driver.ServerID).Return(cloud.ServerAction{}, nil)
	servers.On("GetServerById", driver.ServerID).Return("STOPPED")

	err := driver.Kill()
	servers.AssertCalled(t, "StopServer", driver.ServerID)
	assert.NoError(t, err)
}

func TestGetURL(t *testing.T) {
	driver, client := newTestDriver()
	servers := new(mockServersAPI)
	client.On("Servers").Return(servers)
	servers.On("GetServerById", driver.ServerID).Return("RUNNING")

	driver.IPAddress = "10.0.0.10"
	url, err := driver.GetURL()
	assert.NotEmpty(t, url)
	assert.NoError(t, err)
}

func TestGetState(t *testing.T) {
	driver, client := newTestDriver()
	servers := new(mockServersAPI)
	client.On("Servers").Return(servers)
	states := map[state.State][]string{
		state.Starting: {
			"NEW",
			"BUILDING",
			"CONFIGURING_NETWORK",
			"VERIFYING",
			"STARTING",
		},
		state.Running:  {"RUNNING"},
		state.None:     {"DESTROYED"},
		state.Stopping: {"STOPPING", "DESTROYING"},
		state.Stopped:  {"STOPPED"},
		state.Error:    {"ERROR"},
	}
	for s, statuses := range states {
		for _, status := range statuses {
			servers.On("GetServerById", driver.ServerID).Return(status).Once()
			state, err := driver.GetState()
			assert.Equal(t, s, state)
			assert.NoError(t, err)
		}
	}
}

func TestRemove(t *testing.T) {
	driver, client := newTestDriver()
	servers := new(mockServersAPI)
	client.On("Servers").Return(servers)
	servers.On("DestroyServer", driver.ServerID).Return(cloud.ServerAction{}, nil)

	err := driver.Remove()
	servers.AssertCalled(t, "DestroyServer", driver.ServerID)
	assert.NoError(t, err)
}

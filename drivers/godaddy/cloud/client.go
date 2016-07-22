package cloud

type Client struct {
	sshKeys *SshKeysApi
	servers *ServersApi
}

type ServersClient interface {
	AddServer(body ServerCreate) (Server, error)
	GetServerById(serverId string) (Server, error)
	StartServer(serverId string) (ServerAction, error)
	StopServer(serverId string) (ServerAction, error)
	DestroyServer(serverId string) (ServerAction, error)
}

type ClientWrapper interface {
	SSHKeys() *SshKeysApi
	Servers() ServersClient
}

func NewClient(basePath, apiKey string) ClientWrapper {
	return &Client{
		sshKeys: NewSshKeysApi(basePath, apiKey),
		servers: NewServersApi(basePath, apiKey),
	}
}

func (c *Client) SSHKeys() *SshKeysApi {
	return c.sshKeys
}

func (c *Client) Servers() ServersClient {
	return c.servers
}

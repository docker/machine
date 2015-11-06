package rpcdriver

import (
	"fmt"
	"net/rpc"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin/localbinary"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/mcnflag"
	"github.com/docker/machine/libmachine/state"
	"github.com/docker/machine/libmachine/version"
)

var (
	heartbeatInterval = 200 * time.Millisecond
)

type RPCClientDriver struct {
	plugin          localbinary.DriverPlugin
	heartbeatDoneCh chan bool
	Client          *InternalClient
}

type RPCCall struct {
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
}

type InternalClient struct {
	MachineName string
	RPCClient   *rpc.Client
}

func (ic *InternalClient) Call(serviceMethod string, args interface{}, reply interface{}) error {
	if serviceMethod != "RPCServerDriver.Heartbeat" {
		log.Debugf("(%s) Calling %+v", ic.MachineName, serviceMethod)
	}
	return ic.RPCClient.Call(serviceMethod, args, reply)
}

func NewInternalClient(rpcclient *rpc.Client) *InternalClient {
	return &InternalClient{
		RPCClient: rpcclient,
	}
}

func NewRPCClientDriver(rawDriverData []byte, driverName string) (*RPCClientDriver, error) {
	mcnName := ""

	p, err := localbinary.NewPlugin(driverName)
	if err != nil {
		return nil, err
	}

	go func() {
		if err := p.Serve(); err != nil {
			// TODO: Is this best approach?
			log.Warn(err)
			return
		}
	}()

	addr, err := p.Address()
	if err != nil {
		return nil, fmt.Errorf("Error attempting to get plugin server address for RPC: %s", err)
	}

	rpcclient, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		return nil, err
	}

	c := &RPCClientDriver{
		Client:          NewInternalClient(rpcclient),
		heartbeatDoneCh: make(chan bool),
	}

	go func(c *RPCClientDriver) {
		for {
			select {
			case <-c.heartbeatDoneCh:
				return
			default:
				if err := c.Client.Call("RPCServerDriver.Heartbeat", struct{}{}, nil); err != nil {
					log.Warnf("Error attempting heartbeat call to plugin server: %s", err)
					c.Close()
					return
				}
				time.Sleep(heartbeatInterval)
			}
		}
	}(c)

	var serverVersion int
	if err := c.Client.Call("RPCServerDriver.GetVersion", struct{}{}, &serverVersion); err != nil {
		return nil, err
	}

	if serverVersion != version.APIVersion {
		return nil, fmt.Errorf("Driver binary uses an incompatible API version (%d)", serverVersion)
	}
	log.Debug("Using API Version ", serverVersion)

	if err := c.SetConfigRaw(rawDriverData); err != nil {
		return nil, err
	}

	mcnName = c.GetMachineName()
	p.MachineName = mcnName
	c.Client.MachineName = mcnName
	c.plugin = p

	return c, nil
}

func (c *RPCClientDriver) MarshalJSON() ([]byte, error) {
	return c.GetConfigRaw()
}

func (c *RPCClientDriver) UnmarshalJSON(data []byte) error {
	return c.SetConfigRaw(data)
}

func (c *RPCClientDriver) Close() error {
	c.heartbeatDoneCh <- true
	close(c.heartbeatDoneCh)

	log.Debug("Making call to close connection to plugin binary")

	if err := c.plugin.Close(); err != nil {
		return err
	}

	log.Debug("Making call to close driver server")

	if err := c.Client.Call("RPCServerDriver.Close", struct{}{}, nil); err != nil {
		return err
	}

	log.Debug("Successfully made call to close driver server")

	return nil
}

// Helper method to make requests which take no arguments and return simply a
// string, e.g. "GetIP".
func (c *RPCClientDriver) rpcStringCall(method string) (string, error) {
	var info string

	if err := c.Client.Call(method, struct{}{}, &info); err != nil {
		return "", err
	}

	return info, nil
}

func (c *RPCClientDriver) GetCreateFlags() []mcnflag.Flag {
	var flags []mcnflag.Flag

	if err := c.Client.Call("RPCServerDriver.GetCreateFlags", struct{}{}, &flags); err != nil {
		log.Warnf("Error attempting call to get create flags: %s", err)
	}

	return flags
}

func (c *RPCClientDriver) SetConfigRaw(data []byte) error {
	return c.Client.Call("RPCServerDriver.SetConfigRaw", data, nil)
}

func (c *RPCClientDriver) GetConfigRaw() ([]byte, error) {
	var data []byte

	if err := c.Client.Call("RPCServerDriver.GetConfigRaw", struct{}{}, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// DriverName returns the name of the driver
func (c *RPCClientDriver) DriverName() string {
	driverName, err := c.rpcStringCall("RPCServerDriver.DriverName")
	if err != nil {
		log.Warnf("Error attempting call to get driver name: %s", err)
	}

	return driverName
}

func (c *RPCClientDriver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	return c.Client.Call("RPCServerDriver.SetConfigFromFlags", &flags, nil)
}

func (c *RPCClientDriver) GetURL() (string, error) {
	return c.rpcStringCall("RPCServerDriver.GetURL")
}

func (c *RPCClientDriver) GetMachineName() string {
	name, err := c.rpcStringCall("RPCServerDriver.GetMachineName")
	if err != nil {
		log.Warnf("Error attempting call to get machine name: %s", err)
	}

	return name
}

func (c *RPCClientDriver) GetIP() (string, error) {
	return c.rpcStringCall("RPCServerDriver.GetIP")
}

func (c *RPCClientDriver) GetSSHHostname() (string, error) {
	return c.rpcStringCall("RPCServerDriver.GetSSHHostname")
}

// GetSSHKeyPath returns the key path
// TODO:  This method doesn't even make sense to have with RPC.
func (c *RPCClientDriver) GetSSHKeyPath() string {
	path, err := c.rpcStringCall("RPCServerDriver.GetSSHKeyPath")
	if err != nil {
		log.Warnf("Error attempting call to get SSH key path: %s", err)
	}

	return path
}

func (c *RPCClientDriver) GetSSHPort() (int, error) {
	var port int

	if err := c.Client.Call("RPCServerDriver.GetSSHPort", struct{}{}, &port); err != nil {
		return 0, err
	}

	return port, nil
}

func (c *RPCClientDriver) GetSSHUsername() string {
	username, err := c.rpcStringCall("RPCServerDriver.GetSSHUsername")
	if err != nil {
		log.Warnf("Error attempting call to get SSH username: %s", err)
	}

	return username
}

func (c *RPCClientDriver) GetState() (state.State, error) {
	var s state.State

	if err := c.Client.Call("RPCServerDriver.GetState", struct{}{}, &s); err != nil {
		return state.Error, err
	}

	return s, nil
}

func (c *RPCClientDriver) PreCreateCheck() error {
	return c.Client.Call("RPCServerDriver.PreCreateCheck", struct{}{}, nil)
}

func (c *RPCClientDriver) Create() error {
	return c.Client.Call("RPCServerDriver.Create", struct{}{}, nil)
}

func (c *RPCClientDriver) Remove() error {
	return c.Client.Call("RPCServerDriver.Remove", struct{}{}, nil)
}

func (c *RPCClientDriver) Start() error {
	return c.Client.Call("RPCServerDriver.Start", struct{}{}, nil)
}

func (c *RPCClientDriver) Stop() error {
	return c.Client.Call("RPCServerDriver.Stop", struct{}{}, nil)
}

func (c *RPCClientDriver) Restart() error {
	return c.Client.Call("RPCServerDriver.Restart", struct{}{}, nil)
}

func (c *RPCClientDriver) Kill() error {
	return c.Client.Call("RPCServerDriver.Kill", struct{}{}, nil)
}

func (c *RPCClientDriver) LocalArtifactPath(file string) string {
	var path string

	if err := c.Client.Call("RPCServerDriver.LocalArtifactPath", file, &path); err != nil {
		log.Warnf("Error attempting call to get LocalArtifactPath: %s", err)
	}

	return path
}

func (c *RPCClientDriver) GlobalArtifactPath() string {
	globalArtifactPath, err := c.rpcStringCall("RPCServerDriver.GlobalArtifactPath")
	if err != nil {
		log.Warnf("Error attempting call to get GlobalArtifactPath: %s", err)
	}

	return globalArtifactPath
}

func (c *RPCClientDriver) Upgrade() error {
	return c.Client.Call("RPCServerDriver.Upgrade", struct{}{}, nil)
}

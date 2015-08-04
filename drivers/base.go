package drivers

import "path/filepath"

// BaseDriver - Embed this struct into drivers to provide the common set
// of fields and functions.
type BaseDriver struct {
	storePath      string
	IPAddress      string
	SSHUser        string
	SSHPort        int
	MachineName    string
	CaCertPath     string
	PrivateKeyPath string
	SwarmMaster    bool
	SwarmHost      string
	SwarmDiscovery string
}

// NewBaseDriver - Get an instance of a BaseDriver
func NewBaseDriver(machineName string, storePath string, caCert string, privateKey string) *BaseDriver {
	return &BaseDriver{
		MachineName:    machineName,
		storePath:      storePath,
		CaCertPath:     caCert,
		PrivateKeyPath: privateKey,
	}
}

// GetSSHKeyPath -
func (d *BaseDriver) GetSSHKeyPath() string {
	return d.ResolveStorePath("id_rsa")
}

// ResolveStorePath -
func (d *BaseDriver) ResolveStorePath(path string) string {
	return filepath.Join(d.storePath, path)
}

// AuthorizePort -
func (d *BaseDriver) AuthorizePort(ports []*Port) error {
	return nil
}

// DeauthorizePort -
func (d *BaseDriver) DeauthorizePort(ports []*Port) error {
	return nil
}

// DriverName - This must be implemented in every driver
func (d *BaseDriver) DriverName() string {
	return "unknown"
}

// GetMachineName -
func (d *BaseDriver) GetMachineName() string {
	return d.MachineName
}

// GetSSHPort -
func (d *BaseDriver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = 22
	}

	return d.SSHPort, nil
}

// GetSSHUsername -
func (d *BaseDriver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "root"
	}

	return d.SSHUser
}

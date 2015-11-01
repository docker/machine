package drivers

import "path/filepath"

// BaseDriver - Embed this struct into drivers to provide the common set
// of fields and functions.
type BaseDriver struct {
	IPAddress      string
	SSHUser        string
	SSHPort        int
	MachineName    string
	SwarmMaster    bool
	SwarmHost      string
	SwarmDiscovery string
	StorePath      string
}

// GetSSHKeyPath -
func (d *BaseDriver) GetSSHKeyPath() string {
	return filepath.Join(d.StorePath, "machines", d.MachineName, "id_rsa")
}

// DriverName returns the name of the driver
func (d *BaseDriver) DriverName() string {
	return "unknown"
}

// GetIP returns the ip
func (d *BaseDriver) GetMachineName() string {
	return d.MachineName
}

// ResolveStorePath -
func (d *BaseDriver) ResolveStorePath(file string) string {
	return filepath.Join(d.StorePath, "machines", d.MachineName, file)
}

// GetSSHPort returns the ssh port, 22 if not specified
func (d *BaseDriver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = 22
	}

	return d.SSHPort, nil
}

// GetSSHUsername returns the ssh user name, root if not specified
func (d *BaseDriver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "root"
	}

	return d.SSHUser
}

// PreCreateCheck is called to enforce pre-creation steps
func (d *BaseDriver) PreCreateCheck() error {
	return nil
}

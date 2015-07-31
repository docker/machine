/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package vmwareappcatalyst

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/log"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

const (
	insecureSSHKey = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`
)

// Driver for VMware AppCatalyst
type Driver struct {
	MachineName    string
	IPAddress      string
	Memory         int
	DiskSize       int
	CPU            int
	CaCertPath     string
	PrivateKeyPath string
	SwarmMaster    bool
	SwarmHost      string
	SwarmDiscovery string
	CPUS           int
	SSHUser        string
	SSHPort        int
	VMLocation     string
	APIPort        int

	storePath string
}

func init() {
	drivers.Register("vmwareappcatalyst", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// GetCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.IntFlag{
			EnvVar: "APPCATALYST_CPU_COUNT",
			Name:   "vmwareappcatalyst-cpu-count",
			Usage:  "number of CPUs for the machine (-1 to use the number of CPUs available)",
			Value:  1,
		},
		cli.IntFlag{
			EnvVar: "APPCATALYST_MEMORY_SIZE",
			Name:   "vmwareappcatalyst-memory-size",
			Usage:  "AppCatalyst size of memory for host VM (in MB)",
			Value:  1024,
		},
		cli.StringFlag{
			EnvVar: "APPCATALYST_VM_LOCATION",
			Name:   "vmwareappcatalyst-vm-location",
			Usage:  "Location of AppCatalyst VMs",
			Value:  os.Getenv("HOME") + "/Documents/AppCatalyst",
		},
		cli.IntFlag{
			EnvVar: "APPCATALYST_API_PORT",
			Name:   "vmwareappcatalyst-api-port",
			Usage:  "AppCatalyst REST API port",
			Value:  8080,
		},
	}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	return &Driver{MachineName: machineName, storePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey}, nil
}

func (d *Driver) AuthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) DeauthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) GetSSHHostname() (string, error) {
	return d.GetIP()
}

func (d *Driver) GetSSHKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) GetSSHPort() (int, error) {
	if d.SSHPort == 0 {
		d.SSHPort = 22
	}

	return d.SSHPort, nil
}

func (d *Driver) GetSSHUsername() string {
	if d.SSHUser == "" {
		d.SSHUser = "photon"
	}

	return d.SSHUser
}

func (d *Driver) DriverName() string {
	return "vmwareappcatalyst"
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.Memory = flags.Int("vmwareappcatalyst-memory-size")
	d.CPU = flags.Int("vmwareappcatalyst-cpu-count")
	d.SwarmMaster = flags.Bool("swarm-master")
	d.SwarmHost = flags.String("swarm-host")
	d.SwarmDiscovery = flags.String("swarm-discovery")
	d.SSHUser = "photon"
	d.SSHPort = 22
	d.VMLocation = flags.String("vmwareappcatalyst-vm-location")
	d.APIPort = flags.Int("vmwareappcatalyst-api-port")

	// We support a maximum of 16 cpu.
	if d.CPU > 16 {
		d.CPU = 16
	}

	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	if ip == "" {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	s, err := d.GetState()
	if err != nil {
		return "", err
	}
	if s != state.Running {
		return "", drivers.ErrHostIsNotRunning
	}

	c, err := NewClient("http://localhost:" + strconv.Itoa(d.APIPort))
	if err != nil {
		return "", err
	}

	var ip string
	for i := 1; i <= 60; i++ {
		vmip, err := c.GetVMIPAddress(d.MachineName)
		if err != nil && vmip.Code != 200 {
			log.Debugf("Not there yet %d/%d, error code: %d, message %s", i, 60, vmip.Code, err)
			time.Sleep(2 * time.Second)
			continue
		}
		ip = vmip.Message
		log.Debugf("Got an ip: %s", ip)
		break
	}

	if ip == "" {
		return "", fmt.Errorf("machine didn't return an IP after 120 seconds, aborting")
	}

	return ip, nil
}

func (d *Driver) GetState() (state.State, error) {
	c, err := NewClient("http://localhost:" + strconv.Itoa(d.APIPort))
	if err != nil {
		return state.Error, err
	}

	power, err := c.GetPowerVM(d.MachineName)
	if err != nil {
		return state.Error, err
	}

	switch power.Message {
	case "powering on":
		return state.Running, nil
	case "powered on":
		return state.Running, nil
	case "powering off":
		return state.Stopped, nil
	case "powered off":
		return state.Stopped, nil
	case "reseting":
		return state.Stopped, nil
	case "suspended":
		return state.Stopped, nil
	case "suspending":
		return state.Stopped, nil
	case "tools_running":
		return state.Running, nil
	case "blocked on msg":
		return state.Error, nil
	}

	// No state whatsoever, errors out
	return state.Error, nil
}

func (d *Driver) PreCreateCheck() error {
	return nil
}

func (d *Driver) Create() error {

	log.Infof("Creating SSH key...")
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return err
	}

	c, err := NewClient("http://localhost:" + strconv.Itoa(d.APIPort))
	if err != nil {
		return err
	}

	log.Infof("Creating VM...")
	_, err = c.CloneVM(d.MachineName, d.MachineName, "photon")
	if err != nil {
		return err
	}

	// Set memory amount
	if err = d.setVMXValue("memsize", strconv.Itoa(d.Memory)); err != nil {
		return err
	}

	// Set cpu number
	if err = d.setVMXValue("numvcpus", strconv.Itoa(d.CPU)); err != nil {
		return err
	}

	log.Infof("Starting %s...", d.MachineName)

	_, err = c.PowerVM(d.MachineName, "on")
	if err != nil {
		return err
	}

	if _, err := d.GetIP(); err != nil {
		return err
	}

	// Add SSH Key to photon user
	log.Infof("Replacing insecure SSH key...")
	if err := d.replaceInsecureSSHKey(); err != nil {
		return err
	}

	var shareName, shareDir string // TODO configurable at some point
	switch runtime.GOOS {
	case "darwin":
		shareName = "Users"
		shareDir = "/Users"
		// TODO "linux" and "windows"
	}

	if shareDir != "" {
		if _, err := os.Stat(shareDir); err != nil && !os.IsNotExist(err) {
			return err
		} else if !os.IsNotExist(err) {
			// add shared folder, create mountpoint and mount it.

			// Enable Shared Folders
			if _, err := c.SetVMSharedFolders(d.MachineName, "true"); err != nil {
				return err
			}

			// Add shared folder to VM
			if _, err := c.AddVMSharedFolder(d.MachineName, shareName, shareDir, 4); err != nil {
				return err
			}

			// create mountpoint and mount shared folder
			if err := d.mountSharedFolder(shareDir, shareName); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *Driver) Start() error {
	if s, _ := d.GetState(); s == state.Running {
		return fmt.Errorf("VM already running")
	}

	log.Infof("Starting %s...", d.MachineName)

	c, err := NewClient("http://localhost:" + strconv.Itoa(d.APIPort))
	if err != nil {
		return err
	}

	_, err = c.PowerVM(d.MachineName, "on")
	if err != nil {
		return err
	}

	log.Debugf("Mounting Shared Folders...")
	var shareName, shareDir string // TODO configurable at some point
	switch runtime.GOOS {
	case "darwin":
		shareName = "Users"
		shareDir = "/Users"
		// TODO "linux" and "windows"
	}

	if shareDir != "" {
		if _, err := os.Stat(shareDir); err != nil && !os.IsNotExist(err) {
			return err
		} else if !os.IsNotExist(err) {
			// create mountpoint and mount shared folder
			if err := d.mountSharedFolder(shareDir, shareName); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *Driver) Stop() error {
	if s, _ := d.GetState(); s == state.Stopped {
		return fmt.Errorf("VM already stopped")
	}

	log.Infof("Gracefully shutting down %s...", d.MachineName)

	c, err := NewClient("http://localhost:" + strconv.Itoa(d.APIPort))
	if err != nil {
		return err
	}
	_, err = c.PowerVM(d.MachineName, "shutdown")
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Remove() error {

	if s, _ := d.GetState(); s == state.Running {
		if err := d.Kill(); err != nil {
			return fmt.Errorf("Error stopping VM before deletion")
		}
	}
	log.Infof("Deleting %s...", d.MachineName)
	c, err := NewClient("http://localhost:" + strconv.Itoa(d.APIPort))
	if err != nil {
		return err
	}
	if err := c.DeleteVM(d.MachineName); err != nil {
		fmt.Println(err)
	}

	return nil
}

func (d *Driver) Restart() error {
	if s, _ := d.GetState(); s != state.Stopped {
		if err := d.Stop(); err != nil {
			return err
		}
	}
	if err := d.Start(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Kill() error {
	if s, _ := d.GetState(); s == state.Stopped {
		return fmt.Errorf("VM already stopped")
	}

	log.Infof("Forcibly halting %s...", d.MachineName)

	c, err := NewClient("http://localhost:" + strconv.Itoa(d.APIPort))
	if err != nil {
		return err
	}
	_, err = c.PowerVM(d.MachineName, "off")
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) vmxPath() string {
	return path.Join(d.storePath, fmt.Sprintf("%s.vmx", d.MachineName))
}

func (d *Driver) vmdkPath() string {
	return path.Join(d.storePath, fmt.Sprintf("%s.vmdk", d.MachineName))
}

func (d *Driver) publicSSHKeyPath() string {
	return d.GetSSHKeyPath() + ".pub"
}

func (d *Driver) readSSHKey() (string, error) {
	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return "", err
	}

	return string(publicKey), nil
}

func (d *Driver) replaceInsecureSSHKey() error {
	addr, err := d.GetSSHHostname()
	if err != nil {
		return err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return err
	}

	tempssh, err := ioutil.TempFile(os.TempDir(), "appcatalyst-ssh-insecure-key")
	defer os.Remove(tempssh.Name())

	_, err = tempssh.WriteString(insecureSSHKey)
	if err != nil {
		return err
	}

	auth := &ssh.Auth{
		Keys: []string{tempssh.Name()},
	}

	client, err := ssh.NewClient(d.GetSSHUsername(), addr, port, auth)
	if err != nil {
		return err
	}

	pubkey, err := d.readSSHKey()
	if err != nil {
		return err
	}

	command := "echo '" + pubkey + "' > ~/.ssh/authorized_keys"

	log.Debugf("About to run SSH command:\n%s", command)
	output, err := client.Output(command)
	log.Debugf("SSH cmd err, output: %v: %s", err, output)
	return err
}

func (d *Driver) mountSharedFolder(shareDir, shareName string) error {
	// create mountpoint and mount shared folder
	addr, err := d.GetSSHHostname()
	if err != nil {
		return err
	}

	port, err := d.GetSSHPort()
	if err != nil {
		return err
	}

	auth := &ssh.Auth{
		Keys: []string{d.GetSSHKeyPath()},
	}

	client, err := ssh.NewClient(d.GetSSHUsername(), addr, port, auth)
	if err != nil {
		return err
	}

	command := "[ ! -d " + shareDir + " ]&& sudo mkdir " + shareDir + "; sudo mount -t vmhgfs .host:/" + shareName + " " + shareDir

	log.Debugf("About to run SSH command:\n%s", command)
	output, err := client.Output(command)
	log.Debugf("SSH cmd err, output: %v: %s", err, output)
	return err
}

func (d *Driver) setVMXValue(vmxkey, vmxvalue string) error {

	vmxPath := d.VMLocation + "/" + d.MachineName + "/" + d.MachineName + ".vmx"

	data, err := ioutil.ReadFile(vmxPath)
	if err != nil {
		return err
	}

	results := make(map[string]string)

	lineRe := regexp.MustCompile(`^(.+?)\s*=\s*"(.*?)"\s*$`)

	for _, line := range strings.Split(string(data), "\n") {
		matches := lineRe.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		key := strings.ToLower(matches[1])
		results[key] = matches[2]
	}
	// set the value
	results[vmxkey] = vmxvalue

	var buf bytes.Buffer

	i := 0
	keys := make([]string, len(results))
	for k := range results {
		keys[i] = k
		i++
	}

	sort.Strings(keys)
	for _, k := range keys {
		buf.WriteString(fmt.Sprintf("%s = \"%s\"\n", k, results[k]))
	}

	log.Debugf("Writing VMX to: %s", vmxPath)
	f, err := os.Create(vmxPath)
	if err != nil {
		return err
	}
	defer f.Close()

	var bufwrite bytes.Buffer
	bufwrite.WriteString(buf.String())
	if _, err = io.Copy(f, &bufwrite); err != nil {
		return err
	}
	return nil

}

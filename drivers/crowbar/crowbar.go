package crowbar  // http://github.com/opencrowbar/core
// Apache 2 License 2015 by Rob Hirschfeld for RackN

import (
	"path/filepath"
	"os/exec"
	"errors"
	"fmt"
	"io/ioutil"
	"time"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/provider"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"
)

// we can initialize these here, short runs mean they have to re-checked
var target 		*Deployment
var source 		*Deployment
var node    	*Node
var password 	string

const (
	SOURCE_POOL		= "system"
	TARGET_POOL		= "docker-machines"
	READY_STATE 	= "docker-ready"
	OS_INSTALL 		= "crowbar-installed-node"
	INSTALL_OS		= "ubuntu-14.04"
	DEFAULT_PASS	= "crowbar"
)

type Driver struct {
	// Docker Machine
	MachineName       string
	storePath         string
	CaCertPath        string
	PrivateKeyPath    string
	// Crowbar 
	Node 		string  // determined from node allocation pool
    URL         string
    User        string
	SourcePool	string  // default "system"
	TargetPool	string  // default "docker-machines"
	TargetOS	string  // default ubuntu-14.04
	ReadyState	string  // default "docker-ready"
}

func init() {
	log.Debugf("Initializing Crowbar Driver")
	drivers.Register("crowbar", &drivers.RegisteredDriver{
		New:            NewDriver,
		GetCreateFlags: GetCreateFlags,
	})
}

// RegisterCreateFlags registers the flags this driver adds to
// "docker hosts create"
func GetCreateFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			EnvVar: "OCB_URL",
			Name:   "crowbar-url",
			Usage:  "OpenCrowbar API URL (include port)",
			Value:  "http://192.168.124.10:3000"},
		cli.StringFlag{
			EnvVar: "OCB_USER",
			Name:   "crowbar-user",
			Usage:  "OpenCrowbar username",
			Value:  "crowbar"},
		cli.StringFlag{
			EnvVar: "OCB_PASSWORD",
			Name:   "crowbar-password",
			Usage:  "OpenCrowbar password",
			Value:  "crowbar"}}
}

func NewDriver(machineName string, storePath string, caCert string, privateKey string) (drivers.Driver, error) {
	return &Driver{MachineName: machineName, storePath: storePath, CaCertPath: caCert, PrivateKeyPath: privateKey, 
					SourcePool: SOURCE_POOL, TargetPool: TARGET_POOL, ReadyState: READY_STATE,
					TargetOS: INSTALL_OS}, nil
}

func (d *Driver) CrowbarSession() (error) {
	if session == nil {
		log.Debugf("Starting Crowbar Driver for %s", d.MachineName)
		if password == "" {
			log.Warnf("Missing password!  Assuming default")
			password = DEFAULT_PASS
		}
		var err error
		session, err = NewClient(d.User, password, d.URL)
		if err != nil {
			return fmt.Errorf("Could not start Crowbar Driver: %s", err)
		} else {
			log.Debugf("Started Crowbar Driver for %s", d.MachineName)
		}
	}
	return nil
}

func (d *Driver)InitNode() (err error) {
	d.CrowbarSession()
	if node == nil {
		node = &Node{}
		err = node.get(d.Node)
		if err != nil {
			log.Warnf("Did not find Machine %s assigned to Crowbar Node %s.  Returned error: %s", d.MachineName, d.Node, err)
		}
	}
	return err
}

func (d *Driver)InitDeployments() (err error) {
	d.CrowbarSession()
	if target == nil {
		target = &Deployment{}
		err = target.get(d.TargetPool)
		if err != nil {
			log.Warnf("Adding Target %s Deployment %s", d.TargetPool, session.URL)
			err = target.add(&NewDeployment{Name: d.TargetPool, Description: "Added for Docker Machine"})
			if err != nil {
				return
			}
		}
		target.commit()
	}
	if source == nil {
		source = &Deployment{}
		err = source.get(d.SourcePool)
		if err != nil {
			log.Warnf("Adding Source %s Deployment %s", d.SourcePool, session.URL)
			err = source.add(&NewDeployment{Name: d.SourcePool, Description: "Added for Docker Machine"})
			if err != nil {
				return
			}
		}
	}
	return
}

func (d *Driver) DriverName() string {
	return "crowbar"
}

func (d *Driver) AuthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) DeauthorizePort(ports []*drivers.Port) error {
	return nil
}

func (d *Driver) SetConfigFromFlags(flags drivers.DriverOptions) error {
	d.URL = flags.String("crowbar-url")
	d.User = flags.String("crowbar-user")
	password = flags.String("crowbar-password")
	log.Debugf("Configurating Crowbar Flags: URL %s User %s", d.URL, d.User)
	return nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetIP() (string, error) {
	d.InitNode()
	addresses, err := node.address("admin")
	// assume that this gets IPv4 - IPv6 is the first returned
	addr := addresses[1]
	sz := len(addr)
	log.Debugf("Crowbar node %s has IP addresses %v", node.Name, addresses)
	return addr[:sz-3], err
}

func (d *Driver) GetMachineName() string {
	return d.MachineName
}

func (d *Driver) GetProviderType() provider.ProviderType {
	return provider.Remote
}

func (d *Driver) GetSSHHostname() (string, error) {
	d.InitNode()
	return d.GetIP()
}

func (d *Driver) GetSSHKeyPath() string {
	return filepath.Join(d.storePath, "id_rsa")
}

func (d *Driver) GetSSHPort() (int, error) {
	return 22, nil
}

func (d *Driver) GetSSHUsername() string {
	return "root"
}

func (d *Driver) GetState() (state.State, error) {
	// to ensure refresheness, get a new node object
	n := &Node{}
	err := n.get(d.Node)
	if err != nil {
		log.Errorf("Crowbar Node %s error getting state: %v", d.Node, err)
		return state.Error, err
	}
	if n.Alive {
			// power on, need to look at target role
			nr, err1 := n.role(d.ReadyState)
			if nr.NodeError || err1 != nil {
				return state.Error, err1
			} 
			switch nr.State {
				case -1:
					return state.Error, err1
				case 0:
					return state.Running, err1				
			}
			return state.Starting, err1
		} else {
			// not alive
			return state.Stopped, err
		}
	return state.None, err
}


func (d *Driver) PreCreateCheck() (error) {
	d.InitDeployments()
	candidates, err := source.nodes()
	if err != nil {
		log.Errorf("Error retrieving candidate nodes from %s: %s", session.URL, err)
		return err
	} else if len(candidates) == 0 {
		log.Warnf("No machines in %s deployment %s", session.URL, source.Name)
	} else {
		log.Debugf("Found %d machines in %s deployment %s (some may not be available)", len(candidates), session.URL, source.Name)
	}
	return err
}

func (d *Driver) Create() (err error) {
	log.Debugf("Crowbar allocating (aka creating) %s using %s", d.MachineName, session.URL)
	d.InitDeployments()
	candidates, _ := source.nodes()
	for i := 0; i<len(candidates) && node == nil; i+=1{
		candidate := candidates[i]
 		//log.Debugf("Crowbar inspect %d node %s: !Admin %v !System &v Avail &v", i, candidate.Name, candidate.ID, !candidate.Admin, !candidate.System, candidate.Available)
 		if !candidate.Admin && !candidate.System && candidate.Available {
 			node = &candidate
 		}
	}
	if node == nil {
		log.Errorf("No available machines in %s deployment %s", session.URL, source.Name)
		return errors.New("No machines")
	}
	log.Debugf("Crowbar picked node %s (%d)", node.Name, node.ID)
	node.Description = "Docker-Machine " + d.MachineName
	node.DeploymentID = target.ID
	err = node.update()
	if err != nil {
		return err
	}
	d.Node = node.Name
	log.Infof("Crowbar assigned node %s from %s to %s (%d)", node.Name, d.SourcePool, d.TargetPool, node.DeploymentID)
	// start setting values
	node.propose()
	// add os install target
	osrole := &Role{}
	osrole.get(OS_INSTALL)
	osinstall := &NodeRole{DeploymentID: target.ID, NodeID: node.ID, RoleID: osrole.ID, Order: 1000 }
	osinstall.add()
	// add ready state target
	role := &Role{}
	role.get(d.ReadyState)
	ready := &NodeRole{DeploymentID: target.ID, NodeID: node.ID, RoleID: role.ID, Order: 2000 }
	ready.add()
	// inject key for docker
	sshkey, err := d.createKeyPair()
	node.addSSHkey(1,sshkey)
	log.Debugf("Crowbar added public key [%s...] to node %s", sshkey[:25], node.Name)
	// set operating system
	log.Debugf("Crowbar set %s operating system to %s", node.Name, d.TargetOS)
	if osAvailable(d.TargetOS) {
		node.setOS(d.TargetOS)
	} else {
		return errors.New("Requested operating system has not been configured in Crowbar")
	}
	// start processing
	node.commit()
	node.refresh()
	log.Debugf("Crowbar node %s Ready State target set to %s (%d)", node.Name, role.Name, role.ID)
	// Crowbar reuses servers, so we need to cleanups known hosts
	ip, _ := d.GetIP()
	log.Debugf("Attempting to removing key for %s to prevent known_hosts MitM failure", ip)
	cmd := "ssh-keygen"
	args := []string{"-f", "~/.ssh/known_hosts", "-R", ip}
	if err_kh := exec.Command(cmd, args...).Run(); err_kh != nil {
		log.Infof("You may need to cleanup your known_hosts file: \"ssh-keygen -f ~/.ssh/known_hosts -R %s\" (Automatic attempt returned: %v)", ip, err_kh)
	}
	log.Infof("Crowbar preparing node %s (process may take several minutes)", node.Name)
	// if we are not ready, then loop a bit
	for s, i := state.None, 0; i<40 && s != state.Running; s, err = d.GetState() {
		log.Debugf("Crowbar waiting for %s machine (state %v)", node.Name, s)
		time.Sleep(time.Second*15)
		if err != nil {
			return err
		} else if s == state.Error {
			return errors.New("node state reported error")
		}
		i += 1
	}
	return err
}

func (d *Driver) createKeyPair() (string, error) {

	file := d.GetSSHKeyPath() + ".pub"
	log.Debugf("Crowbar creating key pair at %s", file)
	if err := ssh.GenerateSSHKey(d.GetSSHKeyPath()); err != nil {
		return "", err
	}
	pk, err := ioutil.ReadFile(file)
	publicKey := string(pk[:])
	if err != nil {
		return "", err
	}
	return publicKey, nil
}

func (d *Driver) Remove() error {
	err := d.InitNode()
	d.InitDeployments()
	log.Debugf("Crowbar releasing node %s back to deployment %s", node.Name, source.Name)
	node.propose()
	node.Available = true
	node.Description = "Released from Docker-Machine"
	node.DeploymentID = source.ID
	err = node.update() 
	node.addSSHkey(1,"key-removed")
	node.redeploy()
	node.commit()
	log.Debugf("Crowbar started redeploy request for node %s", node.Name)
	return err
}

func (d *Driver) Start() error {
	log.Debugf("Crowbar starting node")
	err := d.InitNode()
	node.power("on", "on")
	return err
}

func (d *Driver) Stop() error {
	log.Debugf("Crowbar stopping node")
	err := d.InitNode()
	node.power("off", "reboot")
	return err
}

func (d *Driver) Restart() error {
	log.Debugf("Crowbar restarting node")
	err := d.InitNode()
	node.power("reboot", "reset")
	return err
}

func (d *Driver) Kill() error {
	log.Debugf("Crowbar killing node")
	err := d.InitNode()
	node.power("halt", "reboot")
	return err
}

func (d *Driver) Upgrade() error {
	log.Debugf("Crowbar upgrading node - NOT IMPLEMENTED")
	err := d.InitNode()
	return err
}

func (d *Driver) StartDocker() error {
	return nil
}

func (d *Driver) StopDocker() error {
	return nil
}

func (d *Driver) GetDockerConfigDir() string {
	log.Debugf("Docker Dir")
	return "/etc/docker"
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	log.Debugf("Get SSH Command")
	return &exec.Cmd{}, nil
}

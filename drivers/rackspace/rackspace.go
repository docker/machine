package rackspace

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/utils"
	"github.com/docker/machine/drivers"
	"github.com/docker/machine/ssh"
	"github.com/docker/machine/state"

	"github.com/racker/perigee"
	"github.com/rackspace/gophercloud"
	oskey "github.com/rackspace/gophercloud/openstack/compute/v2/extensions/keypairs"
	osflavors "github.com/rackspace/gophercloud/openstack/compute/v2/flavors"
	osimages "github.com/rackspace/gophercloud/openstack/compute/v2/images"
	osservers "github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"github.com/rackspace/gophercloud/pagination"
	"github.com/rackspace/gophercloud/rackspace"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/flavors"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/images"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/keypairs"
	"github.com/rackspace/gophercloud/rackspace/compute/v2/servers"
)

type Driver struct {
	Username string
	APIKey   string
	Region   string
	ImageID  string
	FlavorID string

	storePath      string
	imageQuery     string
	flavorQuery    string
	KeyPairName    string
	ServerName     string
	ServerID       string
	ServerIPAddr   string
	ServerUsername string
}

type CreateFlags struct {
	Username    *string
	APIKey      *string
	Region      *string
	ImageQuery  *string
	FlavorQuery *string
	ServerName  *string
}

func init() {
	drivers.Register("rackspace", &drivers.RegisteredDriver{
		New:                 NewDriver,
		RegisterCreateFlags: RegisterCreateFlags,
	})
}

func errMissingOption(flagName string) error {
	return fmt.Errorf("rackspace driver requires the --rackspace-%s option", flagName)
}

func combineErrors(errs []error) error {
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		messages := make([]string, len(errs))
		for i, err := range errs {
			messages[i] = err.Error()
		}
		combined := strings.Join(messages, "\n ")
		return fmt.Errorf("Multiple errors encountered:\n %s", combined)
	}
}

func RegisterCreateFlags(cmd *flag.FlagSet) interface{} {
	return &CreateFlags{
		Username: cmd.String(
			[]string{"-rackspace-username"},
			"",
			"Rackspace account username",
		),
		APIKey: cmd.String(
			[]string{"-rackspace-api-key"},
			"",
			"Rackspace API key",
		),
		Region: cmd.String(
			[]string{"-rackspace-region"},
			"",
			"Rackspace region",
		),
		ImageQuery: cmd.String(
			[]string{"-rackspace-image"},
			"0766e5df-d60a-4100-ae8c-07f27ec0148f",
			"Rackspace image ID or name query. Default: Ubuntu 14.10 (Utopic Unicorn) (PVHVM)",
		),
		FlavorQuery: cmd.String(
			[]string{"-rackspace-flavor"},
			"performance1-1",
			"Rackspace flavor ID or name query. Default: 1GB Performance",
		),
		ServerName: cmd.String(
			[]string{"-rackspace-server-name"},
			"",
			"Rackspace server name",
		),
	}
}

func NewDriver(storePath string) (drivers.Driver, error) {
	return &Driver{storePath: storePath}, nil
}

func (d *Driver) DriverName() string {
	return "rackspace"
}

func (d *Driver) SetConfigFromFlags(flagsInterface interface{}) error {
	flags := flagsInterface.(*CreateFlags)
	d.Username = *flags.Username
	d.APIKey = *flags.APIKey
	d.Region = *flags.Region
	d.imageQuery = *flags.ImageQuery
	d.flavorQuery = *flags.FlavorQuery
	d.ServerName = *flags.ServerName

	// Options with derived default values.
	if d.ServerName == "" {
		d.ServerName = "docker-host-" + utils.GenerateRandomID()
	}

	return nil
}

func (d *Driver) Create() error {
	log.Infof("Creating Rackspace server [%s]...", d.ServerName)

	client, err := d.authenticate()
	if err != nil {
		return err
	}

	if err := d.validateForCreate(client); err != nil {
		return err
	}

	if err := d.createSSHKey(client); err != nil {
		return err
	}

	if err := d.createServer(client); err != nil {
		return err
	}

	if err := d.setupDocker(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) GetIP() (string, error) {
	if d.ServerIPAddr == "" {
		return "", errors.New("Server has not been created yet.")
	}
	return d.ServerIPAddr, nil
}

func (d *Driver) GetURL() (string, error) {
	ip, err := d.GetIP()
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("tcp://%s:2376", ip), nil
}

func (d *Driver) GetState() (state.State, error) {
	if d.ServerID == "" {
		return state.None, nil
	}

	client, err := d.authenticate()
	if err != nil {
		return state.None, err
	}

	current, err := servers.Get(client, d.ServerID).Extract()
	if err != nil {
		return state.None, err
	}

	switch current.Status {
	case "BUILD":
		return state.Starting, nil
	case "ACTIVE":
		return state.Running, nil
	case "SUSPENDED":
		return state.Paused, nil
	case "DELETED":
		return state.Stopped, nil
	}

	return state.None, nil
}

func (d *Driver) Start() error {
	return errors.New("Unsupported at this time.")
}

func (d *Driver) Stop() error {
	return errors.New("Unsupported at this time.")
}

func (d *Driver) Remove() error {
	if d.ServerID == "" {
		// Server was not created completely.
		return nil
	}

	client, err := d.authenticate()
	if err != nil {
		return err
	}

	log.Debugf("Deleting the server.")
	if err := servers.Delete(client, d.ServerID).ExtractErr(); err != nil {
		return err
	}

	log.Debugf("Deleting the ssh keypair.")
	if err := keypairs.Delete(client, d.KeyPairName).ExtractErr(); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Restart() error {
	client, err := d.authenticate()
	if err != nil {
		return err
	}

	log.Debugf("Restarting the server.")

	if err := servers.Reboot(client, d.ServerID, osservers.SoftReboot).ExtractErr(); err != nil {
		return err
	}

	log.Debugf("Waiting for server to reboot.")
	if err := servers.WaitForStatus(client, d.ServerID, "ACTIVE", 600); err != nil {
		return err
	}

	return nil
}

func (d *Driver) Upgrade() error {
	return errors.New("Not supported yet")
}

func (d *Driver) Kill() error {
	return d.Remove()
}

func (d *Driver) GetSSHCommand(args ...string) (*exec.Cmd, error) {
	return ssh.GetSSHCommand(d.ServerIPAddr, 22, d.ServerUsername, d.sshKeyPath(), args...), nil
}

func (d *Driver) authenticate() (*gophercloud.ServiceClient, error) {
	if err := d.validateForAuth(); err != nil {
		return nil, err
	}

	log.Debugf("Authenticating with your Rackspace credentials.")

	ao := gophercloud.AuthOptions{
		Username: d.Username,
		APIKey:   d.APIKey,
	}

	providerClient, err := rackspace.AuthenticatedClient(ao)
	if err != nil {
		return nil, err
	}

	serviceClient, err := rackspace.NewComputeV2(providerClient, gophercloud.EndpointOpts{
		Region: d.Region,
	})
	if err != nil {
		return nil, err
	}

	return serviceClient, nil
}

func (d *Driver) validateForAuth() error {
	errs := make([]error, 0)

	// Required options.
	if d.Username == "" {
		errs = append(errs, errMissingOption("username"))
	}
	if d.APIKey == "" {
		errs = append(errs, errMissingOption("api-key"))
	}
	if d.Region == "" {
		errs = append(errs, errMissingOption("region"))
	}

	return combineErrors(errs)
}

func (d *Driver) validateForCreate(client *gophercloud.ServiceClient) error {
	errs := make([]error, 0)

	if imageErr := d.chooseImage(client); imageErr != nil {
		errs = append(errs, imageErr)
	}

	if flavorErr := d.chooseFlavor(client); flavorErr != nil {
		errs = append(errs, flavorErr)
	}

	return combineErrors(errs)
}

func (d *Driver) createSSHKey(client *gophercloud.ServiceClient) error {
	name := d.ServerName + "-key"
	log.Debugf("Creating a new SSH key [%s].", name)

	if err := ssh.GenerateSSHKey(d.sshKeyPath()); err != nil {
		return err
	}

	publicKey, err := ioutil.ReadFile(d.publicSSHKeyPath())
	if err != nil {
		return err
	}

	k, err := keypairs.Create(client, oskey.CreateOpts{
		Name:      name,
		PublicKey: string(publicKey),
	}).Extract()
	if err != nil {
		return err
	}

	d.KeyPairName = k.Name

	return nil
}

func (d *Driver) chooseImage(client *gophercloud.ServiceClient) error {
	setImage := func(im *osimages.Image) {
		log.Debugf("Image '%s' with id=%s chosen.", im.Name, im.ID)
		d.ImageID = im.ID
		if d.ServerUsername == "" && strings.Contains(strings.ToLower(im.Name), "coreos") {
			log.Debugf("")
			d.ServerUsername = "core"
		} else {
			d.ServerUsername = "root"
		}
	}

	if d.imageQuery != "" {
		// First, attempt to interpret the query as an image ID.
		im, err := images.Get(client, d.imageQuery).Extract()
		if err == nil {
			setImage(im)
			return nil
		}

		if casted, ok := err.(*perigee.UnexpectedResponseCodeError); !ok || casted.Actual != 404 {
			return err
		}
	}

	// List the images available and filter by name.
	matchingImages := make([]osimages.Image, 0, 5)
	allImages := make([]osimages.Image, 0, 50)
	lowerQuery := strings.ToLower(d.imageQuery)

	err := images.ListDetail(client, nil).EachPage(func(page pagination.Page) (bool, error) {
		is, err := images.ExtractImages(page)
		if err != nil {
			return false, err
		}

		allImages = append(allImages, is...)
		for _, image := range is {

			if d.imageQuery != "" {
				lowerName := strings.ToLower(image.Name)
				if strings.Contains(lowerName, lowerQuery) {
					matchingImages = append(matchingImages, image)
				}
			}
		}
		return true, nil
	})
	if err != nil {
		return err
	}

	if d.imageQuery == "" {
		// No image query provided. List all available images.
		log.Errorf("You must specify an image for your server.")
		log.Infof("Please choose an image below by providing its ID or name to --rackspace-image:")
		listImages(allImages)
		return fmt.Errorf("Missing required parameter --rackspace-image.")
	}

	switch len(matchingImages) {
	case 1:
		// One match! Use that image.
		match := matchingImages[0]
		setImage(&match)
		return nil
	case 0:
		// No matches. List all available images.
		log.Errorf(`Your image query "%s" didn't match any images.`, d.imageQuery)
		log.Infof("Please choose an image from the following list, by name or by ID.")
		listImages(allImages)
		return fmt.Errorf(`"--rackspace-image %s" didn't match any images.`, d.imageQuery)
	default:
		// Multiple matches. List all matching images.
		log.Errorf(`Your image query "%s" matched %d images.`, d.imageQuery, len(matchingImages))
		log.Infof("These are the choices that matched. Please narrow your search to match only one!")
		listImages(matchingImages)
		return fmt.Errorf(`"--rackspace-image %s" was ambiguous.`, d.imageQuery)
	}
}

func listImages(slice []osimages.Image) {
	maxName, maxID := 0, 0

	for _, image := range slice {
		if len(image.Name) > maxName {
			maxName = len(image.Name)
		}
		if len(image.ID) > maxID {
			maxID = len(image.ID)
		}
	}

	for _, image := range slice {
		log.Infof(" %[2]*[1]s %-[4]*[3]s", image.ID, maxID, image.Name, maxName)
	}
}

func (d *Driver) chooseFlavor(client *gophercloud.ServiceClient) error {
	if d.flavorQuery != "" {
		// First, attempt to interpret the query as a flavor ID.
		fl, err := flavors.Get(client, d.flavorQuery).Extract()
		if err == nil {
			log.Debugf("Flavor '%s' with id=%s chosen.", fl.Name, fl.ID)
			d.FlavorID = fl.ID
			return nil
		}

		if casted, ok := err.(*perigee.UnexpectedResponseCodeError); !ok || casted.Actual != 404 {
			return err
		}
	}

	// List the flavors available and filter by name.
	matchingFlavors := make([]osflavors.Flavor, 0, 5)
	allFlavors := make([]osflavors.Flavor, 0, 50)
	lowerQuery := strings.ToLower(d.flavorQuery)

	err := flavors.ListDetail(client, nil).EachPage(func(page pagination.Page) (bool, error) {
		fs, err := flavors.ExtractFlavors(page)
		if err != nil {
			return false, err
		}

		allFlavors = append(allFlavors, fs...)
		for _, flavor := range fs {

			if d.flavorQuery != "" {
				lowerName := strings.ToLower(flavor.Name)
				if strings.Contains(lowerName, lowerQuery) {
					matchingFlavors = append(matchingFlavors, flavor)
				}
			}
		}
		return true, nil
	})
	if err != nil {
		return err
	}

	if d.flavorQuery == "" {
		// No flavor query provided. List all available flavors.
		log.Errorf("You must specify a flavor for your server.")
		log.Infof("Please choose a flavor below by providing its ID or name to --rackspace-flavor:")
		listFlavors(allFlavors)
		return fmt.Errorf("Missing required parameter --rackspace-flavor.")
	}

	switch len(matchingFlavors) {
	case 1:
		// One match! Use that flavor.
		match := matchingFlavors[0]
		log.Debugf("Flavor '%s' with id=%s has been chosen.", match.Name, match.ID)
		d.FlavorID = match.ID
		return nil
	case 0:
		// No matches. List all available flavors.
		log.Errorf(`Your flavor query "%s" didn't match any flavors.`, d.flavorQuery)
		log.Infof("Please choose a flavor from the following list, by name or by ID.")
		listFlavors(allFlavors)
		return fmt.Errorf(`"--rackspace-flavor %s" didn't match any flavors.`, d.flavorQuery)
	default:
		// Multiple matches. List all matching flavors.
		log.Errorf(`Your image query "%s" matched %d flavors.`, d.flavorQuery, len(matchingFlavors))
		log.Infof("These are the choices that matched. Please narrow your search to match only one!")
		listFlavors(matchingFlavors)
		return fmt.Errorf(`"--rackspace-flavor %s" was ambiguous.`, d.flavorQuery)
	}
}

func listFlavors(slice []osflavors.Flavor) {
	maxName, maxID := 0, 0

	for _, flavor := range slice {
		if len(flavor.Name) > maxName {
			maxName = len(flavor.Name)
		}
		if len(flavor.ID) > maxID {
			maxID = len(flavor.ID)
		}
	}

	for _, flavor := range slice {
		log.Infof(" %[2]*[1]s %[4]*[3]s", flavor.ID, maxID, flavor.Name, maxName)
	}
}

func (d *Driver) createServer(client *gophercloud.ServiceClient) error {
	log.Debugf("Launching the server.")

	s, err := servers.Create(client, servers.CreateOpts{
		Name:      d.ServerName,
		ImageRef:  d.ImageID,
		FlavorRef: d.FlavorID,
		KeyPair:   d.KeyPairName,
	}).Extract()
	if err != nil {
		return err
	}

	log.Debugf("Waiting for server %s to launch.", d.ServerName)
	if err = servers.WaitForStatus(client, s.ID, "ACTIVE", 300); err != nil {
		return err
	}

	log.Debugf("Getting details for server %s.", d.ServerName)
	details, err := servers.Get(client, s.ID).Extract()
	if err != nil {
		return err
	}
	d.ServerID = details.ID
	d.ServerIPAddr = details.AccessIPv4

	log.Debugf("Server %s is ready at IP address %s.", d.ServerName, d.ServerIPAddr)

	return nil
}

func (d *Driver) setupDocker() error {
	log.Debugf("Setting up Docker.")

	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", d.ServerIPAddr, 22)); err != nil {
		return err
	}

	kinds := make(map[string]func() error)
	kinds["(which apt && which service)"] = func() error { return d.setupDockerUbuntu() }
	kinds["(yum --help)"] = func() error { return d.setupDockerFedora() }
	kinds["(which docker && which systemctl && which update_engine_client)"] = func() error { return d.setupDockerCoreOS() }

	var buildErr error
	installed := false

	for probe, installCmdFunc := range kinds {
		// The &&/|| bit keeps the ssh command from exiting with an unsuccessful status when the
		// probe fails, which would keep us from being able to tell the difference between
		// connection problems and a failed probe.
		probeCmd := fmt.Sprintf(`%s && echo -n "yes" || echo -n "no"`, probe)
		sshCmd, err := d.GetSSHCommand(probeCmd)
		if err != nil {
			buildErr = err
			break
		}
		output, err := sshCmd.Output()
		if err != nil {
			buildErr = err
			break
		}

		if strings.HasSuffix(string(output), "yes") {
			if err := installCmdFunc(); err != nil {
				buildErr = err
			} else {
				installed = true
			}
			break
		}
	}

	if buildErr != nil {
		log.Errorf("Something broke while I was setting up Docker on this host!")
		log.Errorf("Details: %v", buildErr)
	}

	if !installed {
		log.Errorf("I don't know how to set up Docker on this host!")
	}

	if buildErr != nil || !installed {
		log.Infof(`You'll need to log in with "docker hosts ssh" and:`)
		log.Infof(" * Install Docker if necessary")
		log.Infof(" * Configure Docker to listen on all interfaces")
	}

	return nil
}

const systemdInit = `[Unit]
Description=Docker Socket for the API

[Socket]
ListenStream=2376
BindIPv6Only=both
Service=docker.service

[Install]
WantedBy=sockets.target`

func (d *Driver) setupDockerUbuntu() error {
	if err := drivers.AddPublicKeyToAuthorizedHosts(d, "/.docker/authorized-keys.d"); err != nil {
		return err
	}

	return d.sshAll([]string{
		`curl -sSL https://get.docker.com/ | sh`,
		`service docker stop`,
		`curl https://bfirsh.s3.amazonaws.com/docker/docker-1.3.1-dev-identity-auth -o /usr/bin/docker`,
		`echo 'export DOCKER_OPTS="--auth=identity --host=tcp://0.0.0.0:2376"' >> /etc/default/docker`,
		`service docker start`,
	})
}

func (d *Driver) setupDockerFedora() error {
	if err := drivers.AddPublicKeyToAuthorizedHosts(d, "/.docker/authorized-keys.d"); err != nil {
		return err
	}

	const systemdConfig = `OPTIONS=--auth=identity --host=tcp://0.0.0.0:2376 --selinux-enabled`

	return d.sshAll([]string{
		`curl -sSL https://get.docker.com | sh`,
		`systemctl stop docker`,
		`curl https://bfirsh.s3.amazonaws.com/docker/docker-1.3.1-dev-identity-auth -o /usr/bin/docker`,
		fmt.Sprintf(`echo '%s' > /etc/sysconfig/docker`, systemdConfig),
		`setsebool -P docker_connect_any 1`,
		`firewall-cmd --zone=public --add-port=2376/tcp`,
		`firewall-cmd --permanent --zone=public --add-port=2376/tcp`,
		`systemctl start docker`,
	})
}

func (d *Driver) setupDockerCoreOS() error {
	// Ignore the error here because it boots us from ssh.
	d.sshAll([]string{"update_engine_client -update"})

	if err := ssh.WaitForTCP(fmt.Sprintf("%s:%d", d.ServerIPAddr, 22)); err != nil {
		return err
	}

	// HACK: Replace the docker daemon in place.
	cmd, err := d.GetSSHCommand("sudo curl https://bfirsh.s3.amazonaws.com/docker/docker-1.3.1-dev-identity-auth -o /usr/bin/docker")
	if err != nil {
		return err
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	if err := drivers.AddPublicKeyToAuthorizedHosts(d, "/.docker/authorized-keys.d"); err != nil {
		return err
	}

	serviceCmds := []string{
		`systemctl enable docker-tcp.socket`,
		`systemctl stop docker`,
		`systemctl start docker-tcp.socket`,
		`systemctl start docker`,
	}

	return d.sshAll([]string{
		fmt.Sprintf(`sudo sh -c "echo '%s' > /etc/systemd/system/docker-tcp.socket"`, systemdInit),
		fmt.Sprintf(`sudo sh -c "%s"`, strings.Join(serviceCmds, " && ")),
	})
}

func (d *Driver) sshAll(commands []string) error {
	for _, command := range commands {
		sshCmd, err := d.GetSSHCommand(command)
		if err != nil {
			return err
		}
		if err := sshCmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

func (d *Driver) sshKeyPath() string {
	return path.Join(d.storePath, "id_rsa")
}

func (d *Driver) publicSSHKeyPath() string {
	return d.sshKeyPath() + ".pub"
}

package provisiontest

import (
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/engine"
	"github.com/docker/machine/libmachine/provision"
	"github.com/docker/machine/libmachine/provision/pkgaction"
	"github.com/docker/machine/libmachine/provision/serviceaction"
	"github.com/docker/machine/libmachine/swarm"
)

type FakeDetector struct {
	provision.Provisioner
}

func (fd *FakeDetector) DetectProvisioner(d drivers.Driver) (provision.Provisioner, error) {
	return fd.Provisioner, nil
}

type FakeProvisioner struct{}

func NewFakeProvisioner(d drivers.Driver) provision.Provisioner {
	return &FakeProvisioner{}
}

func (fp *FakeProvisioner) SSHCommand(args string) (string, error) {
	return "", nil
}

func (fp *FakeProvisioner) String() string {
	return "fakeprovisioner"
}

func (fp *FakeProvisioner) GenerateDockerOptions(dockerPort int) (*provision.DockerOptions, error) {
	return nil, nil
}

func (fp *FakeProvisioner) GetDockerOptionsDir() string {
	return ""
}

func (fp *FakeProvisioner) GetAuthOptions() auth.Options {
	return auth.Options{}
}

func (fp *FakeProvisioner) Package(name string, action pkgaction.PackageAction) error {
	return nil
}

func (fp *FakeProvisioner) Hostname() (string, error) {
	return "", nil
}

func (fp *FakeProvisioner) SetHostname(hostname string) error {
	return nil
}

func (fp *FakeProvisioner) CompatibleWithHost() bool {
	return true
}

func (fp *FakeProvisioner) Provision(swarmOptions swarm.Options, authOptions auth.Options, engineOptions engine.Options) error {
	return nil
}

func (fp *FakeProvisioner) Service(name string, action serviceaction.ServiceAction) error {
	return nil
}

func (fp *FakeProvisioner) GetDriver() drivers.Driver {
	return nil
}

func (fp *FakeProvisioner) SetOsReleaseInfo(info *provision.OsRelease) {}

func (fp *FakeProvisioner) GetOsReleaseInfo() (*provision.OsRelease, error) {
	return nil, nil
}

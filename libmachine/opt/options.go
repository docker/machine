package mcnopt

import (
	"fmt"
	"os"

	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
)

type Options struct {
	BaseDir        string
	SSHClientType  ssh.ClientType
	GithubAPIToken string
	SSHConfigFile  string
	SocksProxy     string
}

var (
	defaultOptions = &Options{
		SSHClientType: ssh.External,
		BaseDir:       mcndirs.GetBaseDir(),
		SSHConfigFile: "/dev/null",
	}
)

func Opts() *Options {
	return defaultOptions
}

func SetOpts(opts *Options) {
	defaultOptions = opts

	// TODO: Ideally this would not be scattered state across several
	// modules, but rather presented through a uniform interface.
	mcndirs.BaseDir = opts.BaseDir
	mcnutils.GithubAPIToken = opts.GithubAPIToken
	ssh.SetDefaultClient(opts.SSHClientType)

	SetSSHConfigFile(opts.SSHConfigFile)
	SetSocksProxy(opts.SocksProxy)
}

func SetSSHConfigFile(SSHConfigFile string) {
	defaultOptions.SSHConfigFile = SSHConfigFile
	ssh.SetConfigFile(SSHConfigFile)
}

func SetSocksProxy(SocksProxy string) {
	defaultOptions.SocksProxy = SocksProxy
	if SocksProxy == "" {
		os.Unsetenv("ALL_PROXY")
	} else {
		os.Setenv("ALL_PROXY", fmt.Sprintf("socks5://%s", SocksProxy))
	}
}

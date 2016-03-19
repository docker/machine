package mcnopt

import (
	"github.com/docker/machine/commands/mcndirs"
	"github.com/docker/machine/libmachine/mcnutils"
	"github.com/docker/machine/libmachine/ssh"
)

type Options struct {
	BaseDir        string
	SSHClientType  ssh.ClientType
	GithubAPIToken string
}

var (
	defaultOptions = &Options{
		SSHClientType: ssh.External,
		BaseDir:       mcndirs.GetBaseDir(),
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
}

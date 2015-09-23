<!--[metadata]>
+++
title = "env"
description = "Set environment variables on a machine"
keywords = ["machine, env, subcommand"]
[menu.main]
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# env

Set environment variables to dictate that `docker` should run a command against
a particular machine.

`docker-machine env machinename` will print out `export` commands which can be
run in a subshell. Running `docker-machine env -u` will print `unset` commands
which reverse this effect.

```
$ env | grep DOCKER
$ eval "$(docker-machine env dev)"
$ env | grep DOCKER
DOCKER_HOST=tcp://192.168.99.101:2376
DOCKER_CERT_PATH=/Users/nathanleclaire/.docker/machines/.client
DOCKER_TLS_VERIFY=1
DOCKER_MACHINE_NAME=dev
$ # If you run a docker command, now it will run against that host.
$ eval "$(docker-machine env -u)"
$ env | grep DOCKER
$ # The environment variables have been unset.
```

The output described above is intended for the shells `bash` and `zsh` (if
you're not sure which shell you're using, there's a very good possibility that
it's `bash`). However, these are not the only shells which Docker Machine
supports.

If you are using `fish` and the `SHELL` environment variable is correctly set to
the path where `fish` is located, `docker-machine env name` will print out the
values in the format which `fish` expects:

```
set -x DOCKER_TLS_VERIFY 1;
set -x DOCKER_CERT_PATH "/Users/nathanleclaire/.docker/machine/machines/overlay";
set -x DOCKER_HOST tcp://192.168.99.102:2376;
set -x DOCKER_MACHINE_NAME overlay
# Run this command to configure your shell:
# eval "$(docker-machine env overlay)"
```

If you are on Windows and using Powershell or `cmd.exe`, `docker-machine env`
cannot detect your shell automatically, but it does have support for these
shells. In order to use them, specify which shell you would like to print the
options for using the `--shell` flag for `docker-machine env`.

For Powershell:

```
$ docker-machine.exe env --shell powershell dev
$Env:DOCKER_TLS_VERIFY = "1"
$Env:DOCKER_HOST = "tcp://192.168.99.101:2376"
$Env:DOCKER_CERT_PATH = "C:\Users\captain\.docker\machine\machines\dev"
$Env:DOCKER_MACHINE_NAME = "dev"
# Run this command to configure your shell:
# docker-machine.exe env --shell=powershell dev | Invoke-Expression
```

For `cmd.exe`:

```
$ docker-machine.exe env --shell cmd dev
set DOCKER_TLS_VERIFY=1
set DOCKER_HOST=tcp://192.168.99.101:2376
set DOCKER_CERT_PATH=C:\Users\captain\.docker\machine\machines\dev
set DOCKER_MACHINE_NAME=dev
# Run this command to configure your shell: copy and paste the above values into your command prompt
```

## Excluding the created machine from proxies

The env command supports a `--no-proxy` flag which will ensure that the created
machine's IP address is added to the [`NO_PROXY`/`no_proxy` environment
variable](https://wiki.archlinux.org/index.php/Proxy_settings).

This is useful when using `docker-machine` with a local VM provider (e.g.
`virtualbox` or `vmwarefusion`) in network environments where a HTTP proxy is
required for internet access.

```
$ docker-machine env --no-proxy default
export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://192.168.99.104:2376"
export DOCKER_CERT_PATH="/Users/databus23/.docker/machine/certs"
export DOCKER_MACHINE_NAME="default"
export NO_PROXY="192.168.99.104"
# Run this command to configure your shell:
# eval "$(docker-machine env default)"
```

You may also want to visit the [documentation on setting `HTTP_PROXY` for the
created daemon using the `--engine-env` flag for `docker-machine
create`](https://docs.docker.com/machine/reference/create/#specifying-configuration-options-for-the-created-docker-engine).

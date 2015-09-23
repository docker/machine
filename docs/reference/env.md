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

## Proxy exclusion
The env command supports a `--no-proxy` flag that will also add the `DOCKER_HOST` to the `NO_PROXY`/`no_proxy`  environment variable (which ever is already defined).

```
docker-machine env dev --no-proxy
set -x DOCKER_TLS_VERIFY "1";
set -x DOCKER_HOST "tcp://172.16.77.135:2376";
set -x DOCKER_CERT_PATH "/Users/fabus/.docker/machine/machines/dev";
set -x DOCKER_MACHINE_NAME "dev";
set -x NO_PROXY "172.16.77.135";
# Run this command to configure your shell:
# eval (docker-machine env dev)
```

This is useful when using docker-machine with a local VM provider (e.g. virtualbox or vmware fusion/workstation) in network environments where a http proxy is needed for internet access.


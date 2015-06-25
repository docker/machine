<!--[metadata]>
+++
title = "Overview of Docker Machine"
description = "Introduction and Overview of Machine"
keywords = ["docker, machine, amazonec2, azure, digitalocean, google, openstack, rackspace, softlayer, virtualbox, vmwarefusion, vmwarevcloudair, vmwarevsphere, exoscale"]
[menu.main]
parent="smn_workw_machine"
+++
<![end-metadata]-->


# Docker Machine

> **Note**: Machine is currently in beta, so things are likely to change. We
> don't recommend you use it in production yet.

Machine lets you create Docker hosts on your computer, on cloud providers, and
inside your own data center. It creates servers, installs Docker on them, then
configures the Docker client to talk to them.

Once your Docker host has been created, it then has a number of commands for
managing them:

 - Starting, stopping, restarting
 - Upgrading Docker
 - Configuring the Docker client to talk to your host

## Getting help

Docker Machine is still in its infancy and under active development. If you need
help, would like to contribute, or simply want to talk about to the project with
like-minded individuals, we have a number of open channels for communication.

- To report bugs or file feature requests: please use the [issue tracker on
  Github](https://github.com/docker/machine/issues).
- To talk about the project with people in real time: please join the
  `#docker-machine` channel on IRC.
- To contribute code or documentation changes: please [submit a pull request on
  Github](https://github.com/docker/machine/pulls).

For more information and resources, please visit
[https://docs.docker.com/project/get-help/](https://docs.docker.com/project/get-help/).

## Installation

Docker Machine is supported on Windows, OS X, and Linux and is installable as one
standalone binary. The links to the binaries for the various platforms and
architectures are below:

- [Windows - 32bit](https://github.com/docker/machine/releases/download/v0.3.0/docker-machine_windows-386.exe)
- [Windows - 64bit](https://github.com/docker/machine/releases/download/v0.3.0/docker-machine_windows-amd64.exe)
- [OSX - x86_64](https://github.com/docker/machine/releases/download/v0.3.0/docker-machine_darwin-amd64)
- [OSX - (old macs)](https://github.com/docker/machine/releases/download/v0.3.0/docker-machine_darwin-386)
- [Linux - x86_64](https://github.com/docker/machine/releases/download/v0.3.0/docker-machine_linux-amd64)
- [Linux - i386](https://github.com/docker/machine/releases/download/v0.3.0/docker-machine_linux-386)

### OS X and Linux

To install on OS X or Linux, download the proper binary to somewhere in your
`PATH` (e.g. `/usr/local/bin`) and make it executable. For instance, to install on
most OS X machines these commands should suffice:

```
$ curl -L https://github.com/docker/machine/releases/download/v0.3.0/docker-machine_darwin-amd64 > /usr/local/bin/docker-machine
$ chmod +x /usr/local/bin/docker-machine
```

For Linux, just substitute "linux" for "darwin" in the binary name above.

Now you should be able to check the version with `docker-machine -v`:

```
$ docker-machine -v
machine version 0.3.0
```

In order to run Docker commands on your machines without having to use SSH, make
sure to install the Docker client as well, e.g.:

```
$ curl -L https://get.docker.com/builds/Darwin/x86_64/docker-latest > /usr/local/bin/docker
```

### Windows

Currently, Docker recommends that you install and use Docker Machine on Windows
with [msysgit](https://msysgit.github.io/). This will provide you with some
programs that Docker Machine relies on such as `ssh`, as well as a functioning
shell.

When you have installed msysgit, start up the terminal prompt and run the
following commands. Here it is assumed that you are on a 64-bit Windows
installation. If you are on a 32-bit installation, please substitute "i386" for
"x86_64" in the URLs mentioned.

First, install the Docker client binary:

```
$ curl -L https://get.docker.com/builds/Windows/x86_64/docker-latest.exe > /bin/docker
```

Next, install the Docker Machine binary:

```
$ curl -L https://github.com/docker/machine/releases/download/v0.3.0/docker-machine_windows-amd64.exe > /bin/docker-machine
```

Now running `docker-machine` should work.

```
$ docker-machine -v
machine version 0.3.0
```

## Getting started with Docker Machine using a local VM

Let's take a look at using `docker-machine` for creating, using, and managing a
Docker host inside of [VirtualBox](https://www.virtualbox.org/).

First, ensure that
[VirtualBox 4.3.28](https://www.virtualbox.org/wiki/Downloads) is correctly
installed on your system.

If you run the `docker-machine ls` command to show all available machines, you will see
that none have been created so far.

```
$ docker-machine ls
NAME   ACTIVE   DRIVER   STATE   URL
```

To create one, we run the `docker-machine create` command, passing the string
`virtualbox` to the `--driver` flag. The final argument we pass is the name of
the machine - in this case, we will name our machine "dev".

This command will download a lightweight Linux distribution
([boot2docker](https://github.com/boot2docker/boot2docker)) with the Docker
daemon installed, and will create and start a VirtualBox VM with Docker running.


```
$ docker-machine create --driver virtualbox dev
Creating CA: /home/username/.docker/machine/certs/ca.pem
Creating client certificate: /home/username/.docker/machine/certs/cert.pem
Image cache does not exist, creating it at /home/username/.docker/machine/cache...
No default boot2docker iso found locally, downloading the latest release...
Downloading https://github.com/boot2docker/boot2docker/releases/download/v1.6.2/boot2docker.iso to /home/username/.docker/machine/cache/boot2docker.iso...
Creating VirtualBox VM...
Creating SSH key...
Starting VirtualBox VM...
Starting VM...
To see how to connect Docker to this machine, run: docker-machine env dev
```

You can see the machine you have created by running the `docker-machine ls`
command again:

```
$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL                         SWARM
dev             virtualbox   Running   tcp://192.168.99.100:2376
```

Next, as noted in the output of the `docker-machine create` command, we have to
tell Docker to talk to that machine. You can do this with the `docker-machine
env` command. For example,

```
$ eval "$(docker-machine env dev)"
$ docker ps
```

> **Note**: If you are using `fish`, or a Windows shell such as
> Powershell/`cmd.exe` the above method will not work as described. Instead,
> see [the `env` command's documentation](https://docs.docker.com/machine/#env)
> to learn how to set the environment variables for your shell.

This will set environment variables that the Docker client will read which specify
the TLS settings. Note that you will need to do that every time you open a new tab or
restart your machine.

To see what will be set, run `docker-machine env dev`.

```
$ docker-machine env dev
export DOCKER_TLS_VERIFY="1"
export DOCKER_HOST="tcp://172.16.62.130:2376"
export DOCKER_CERT_PATH="/Users/<your username>/.docker/machine/machines/dev"
export DOCKER_MACHINE_NAME="dev"
# Run this command to configure your shell:
# eval "$(docker-machine env dev)"
```

You can now run Docker commands on this host:

```
$ docker run busybox echo hello world
Unable to find image 'busybox' locally
Pulling repository busybox
e72ac664f4f0: Download complete
511136ea3c5a: Download complete
df7546f9f060: Download complete
e433a6c5b276: Download complete
hello world
```

Any exposed ports are available on the Docker host’s IP address, which you can
get using the `docker-machine ip` command:

```
$ docker-machine ip dev
192.168.99.100
```

For instance, you can try running a webserver ([nginx](https://nginx.org)) in a
container with the following command:

```
$ docker run -d -p 8000:80 nginx
```

When the image is finished pulling, you can hit the server at port 8000 on the
IP address given to you by `docker-machine ip`. For instance:

```
$ curl $(docker-machine ip dev):8000
<!DOCTYPE html>
<html>
<head>
<title>Welcome to nginx!</title>
<style>
    body {
        width: 35em;
        margin: 0 auto;
        font-family: Tahoma, Verdana, Arial, sans-serif;
    }
</style>
</head>
<body>
<h1>Welcome to nginx!</h1>
<p>If you see this page, the nginx web server is successfully installed and
working. Further configuration is required.</p>

<p>For online documentation and support please refer to
<a href="http://nginx.org/">nginx.org</a>.<br/>
Commercial support is available at
<a href="http://nginx.com/">nginx.com</a>.</p>

<p><em>Thank you for using nginx.</em></p>
</body>
</html>
```

You can create and manage as many local VMs running Docker as you please- just
run `docker-machine create` again. All created machines will appear in the
output of `docker-machine ls`.

If you are finished using a host for the time being, you can stop it with
`docker-machine stop` and later start it again with `docker-machine start`.
Make sure to specify the machine name as an argument:

```
$ docker-machine stop dev
$ docker-machine start dev
```

## Using Docker Machine with a cloud provider

Creating a local virtual machine running Docker is useful and fun, but it is not
the only thing Docker Machine is capable of. Docker Machine supports several
“drivers” which let you use the same interface to create hosts on many different
cloud or local virtualization platforms. This is accomplished by using the
`docker-machine create` command with the `--driver` flag. Here we will be
demonstrating the [Digital Ocean](https://digitalocean.com) driver (called
`digitalocean`), but there are drivers included for several providers including
Amazon Web Services, Google Compute Engine, and Microsoft Azure.

Usually it is required that you pass account verification credentials for these
providers as flags to `docker-machine create`. These flags are unique for each driver.
For instance, to pass a Digital Ocean access token you use the
`--digitalocean-access-token` flag.

Let's take a look at how to do this.

To generate your access token:

1. Go to the Digital Ocean administrator console and click on "API" in the header.
2. Click on "Generate New Token".
3. Give the token a clever name (e.g. "machine"), make sure the "Write" checkbox
is checked, and click on "Generate Token".
4. Grab the big long hex string that is generated (this is your token) and store
it somewhere safe.

Now, run `docker-machine create` with the `digitalocean` driver and pass your key to
the `--digitalocean-access-token` flag.

Example:

```
$ docker-machine create \
    --driver digitalocean \
    --digitalocean-access-token 0ab77166d407f479c6701652cee3a46830fef88b8199722b87821621736ab2d4 \
    staging
Creating SSH key...
Creating Digital Ocean droplet...
To see how to connect Docker to this machine, run: docker-machine env staging
```

For convenience, `docker-machine` will use sensible defaults for choosing
settings such as the image that the VPS is based on, but they can also be
overridden using their respective flags (e.g. `--digitalocean-image`). This is
useful if, for instance, you want to create a nice large instance with a lot of
memory and CPUs (by default `docker-machine` creates a small VPS). For a full
list of the flags/settings available and their defaults, see the output of
`docker-machine create -h`.

When the creation of a host is initiated, a unique SSH key for accessing the
host (initially for provisioning, then directly later if the user runs the
`docker-machine ssh` command) will be created automatically and stored in the
client's directory in `~/.docker/machines`. After the creation of the SSH key,
Docker will be installed on the remote machine and the daemon will be configured
to accept remote connections over TCP using TLS for authentication. Once this
is finished, the host is ready for connection.

To prepare the Docker client to send commands to the remote server we have
created, we can use the subshell method again:

```
$ eval "$(docker-machine env staging)"
```

From this point, the remote host behaves much like the local host we created in
the last section. If we look at `docker-machine ls`, we'll see it is now the
"active" host, indicated by an asterisk (`*`) in that column:

```
$ docker-machine ls
NAME      ACTIVE   DRIVER         STATE     URL
dev                virtualbox     Running   tcp://192.168.99.103:2376
staging   *        digitalocean   Running   tcp://104.236.50.118:2376
```

To remove a host and all of its containers and images, use `docker-machine rm`:

```
$ docker-machine rm dev staging
$ docker-machine ls
NAME      ACTIVE   DRIVER       STATE     URL
```

## Adding a host without a driver

You can add a host to Docker which only has a URL and no driver. Therefore it
can be used an alias for an existing host so you don’t have to type out the URL
every time you run a Docker command.

```
$ docker-machine create --url=tcp://50.134.234.20:2376 custombox
$ docker-machine ls
NAME        ACTIVE   DRIVER    STATE     URL
custombox   *        none      Running   tcp://50.134.234.20:2376
```

## Using Docker Machine with Docker Swarm
Docker Machine can also provision [Swarm](https://github.com/docker/swarm)
clusters. This can be used with any driver and will be secured with TLS.

First, create a Swarm token. Optionally, you can use another discovery service.
See the Swarm docs for details.

To create the token, first create a Machine. This example will use VirtualBox.

```
$ docker-machine create -d virtualbox local
```

Load the Machine configuration into your shell:

```
$ eval "$(docker-machine env local)"
```

Then run generate the token using the Swarm Docker image:

```
$ docker run swarm create
1257e0f0bbb499b5cd04b4c9bdb2dab3
```
Once you have the token, you can create the cluster.

### Swarm master

Create the Swarm master:

```
docker-machine create \
    -d virtualbox \
    --swarm \
    --swarm-master \
    --swarm-discovery token://<TOKEN-FROM-ABOVE> \
    swarm-master
```

Replace `<TOKEN-FROM-ABOVE>` with your random token.
This will create the Swarm master and add itself as a Swarm node.

### Swarm nodes

Now, create more Swarm nodes:

```
docker-machine create \
    -d virtualbox \
    --swarm \
    --swarm-discovery token://<TOKEN-FROM-ABOVE> \
    swarm-node-00
```

You now have a Swarm cluster across two nodes.
To connect to the Swarm master, use `eval $(docker-machine env --swarm swarm-master)`

For example:

```
$ docker-machine env --swarm swarm-master
export DOCKER_TLS_VERIFY=1
export DOCKER_CERT_PATH="/home/ehazlett/.docker/machines/.client"
export DOCKER_HOST=tcp://192.168.99.100:3376
```

You can load this into your environment using
`eval "$(docker-machine env --swarm swarm-master)"`.

Now you can use the Docker CLI to query:

```
$ docker info
Containers: 1
Nodes: 1
 swarm-master: 192.168.99.100:2376
  └ Containers: 2
  └ Reserved CPUs: 0 / 4
  └ Reserved Memory: 0 B / 999.9 MiB
```

## Subcommands

#### active

See which machine is "active" (a machine is considered active if the
`DOCKER_HOST` environment variable points to it).

```
$ docker-machine ls
NAME      ACTIVE   DRIVER         STATE     URL
dev                virtualbox     Running   tcp://192.168.99.103:2376
staging   *        digitalocean   Running   tcp://104.236.50.118:2376
$ echo $DOCKER_HOST
tcp://104.236.50.118:2376
$ docker-machine active
staging
```

#### create

Create a machine.

```
$ docker-machine create --driver virtualbox dev
Creating CA: /home/username/.docker/machine/certs/ca.pem
Creating client certificate: /home/username/.docker/machine/certs/cert.pem
Image cache does not exist, creating it at /home/username/.docker/machine/cache...
No default boot2docker iso found locally, downloading the latest release...
Downloading https://github.com/boot2docker/boot2docker/releases/download/v1.6.2/boot2docker.iso to /home/username/.docker/machine/cache/boot2docker.iso...
Creating VirtualBox VM...
Creating SSH key...
Starting VirtualBox VM...
Starting VM...
To see how to connect Docker to this machine, run: docker-machine env dev
```

##### Filtering create flags by driver in the help text

You may notice that the `docker-machine create` command has a lot of flags due
to the huge plethora of provider-specific options which are available.

```
$ docker-machine create -h | wc -l
145
```

While it is great to have access to all this information, sometimes you simply
want to get a peek at the subset of flags which are applicable to the driver you
are working with. To that extent, specifying an argument to the `-d` flag will
filter the create flags displayed in the help text to only what is applicable to
that provider:

```
$ docker-machine create -d virtualbox
Usage: docker-machine create [OPTIONS] [arg...]

Create a machine

Options:
   --virtualbox-boot2docker-url                                                                         The URL of the boot2docker image. Defaults to the latest available version [$VIRTUALBOX_BOOT2DOCKER_URL]
   --virtualbox-cpu-count "1"                                                                           number of CPUs for the machine (-1 to use the number of CPUs available) [$VIRTUALBOX_CPU_COUNT]
   --virtualbox-disk-size "20000"                                                                       Size of disk for host in MB [$VIRTUALBOX_DISK_SIZE]
   --virtualbox-import-boot2docker-vm                                                                   The name of a Boot2Docker VM to import
   --virtualbox-memory "1024"                                                                           Size of memory for host in MB [$VIRTUALBOX_MEMORY_SIZE]
   --driver, -d "none"                                                                                  Driver to create machine with. Available drivers: amazonec2, azure, digitalocean, exoscale, google, none, openstack, rackspace, softlayer, virtualbox, vmwarefusion, vmwarevcloudair, vmwarevsphere
   --engine-opt [--engine-opt option --engine-opt option]                                               Specify arbitrary opts to include with the created engine in the form opt=value
   --engine-insecure-registry [--engine-insecure-registry option --engine-insecure-registry option]     Specify insecure registries to allow with the created engine
   --engine-registry-mirror [--engine-registry-mirror option --engine-registry-mirror option]           Specify registry mirrors to use
   --engine-label [--engine-label option --engine-label option]                                         Specify labels for the created engine
   --engine-storage-driver "aufs"                                                                       Specify a storage driver to use with the engine
   --swarm                                                                                              Configure Machine with Swarm
   --swarm-master                                                                                       Configure Machine to be a Swarm master
   --swarm-discovery                                                                                    Discovery service to use with Swarm
   --swarm-host "tcp://0.0.0.0:3376"                                                                    ip/socket to listen on for Swarm master
   --swarm-addr                                                                                         addr to advertise for Swarm (default: detect and use the machine IP)
```

##### Specifying configuration options for the created Docker engine

As part of the process of creation, Docker Machine installs Docker and
configures it with some sensible defaults. For instance, it allows connection
from the outside world over TCP with TLS-based encryption and defaults to AUFS
as the [storage
driver](https://docs.docker.com/reference/commandline/cli/#daemon-storage-driver-option)
when available.

There are several cases where the user might want to set options for the created
Docker engine (also known as the Docker _daemon_) themselves. For example, they
may want to allow connection to a [registry](https://docs.docker.com/registry/)
that they are running themselves using the `--insecure-registry` flag for the
daemon. Docker Machine supports the configuration of such options for the
created engines via the `create` command flags which begin with `--engine`.

Note that Docker Machine simply sets the configured parameters on the daemon
and does not set up any of the "dependencies" for you. For instance, if you
specify that the created daemon should use `btrfs` as a storage driver, you
still must ensure that the proper dependencies are installed, the BTRFS
filesystem has been created, and so on.

The following is an example usage:

```
$ docker-machine create -d virtualbox \
    --engine-label foo=bar \
    --engine-label spam=eggs \
    --engine-storage-driver devicemapper \
    --engine-insecure-registry registry.myco.com \
    foobarmachine
```

This will create a virtual machine running locally in Virtualbox which uses the
`devicemapper` storage backend, has the key-value pairs `foo=bar` and
`spam=eggs` as labels on the engine, and allows pushing / pulling from the
insecure registry located at `registry.myco.com`. You can verify much of this
by inspecting the output of `docker info`:

```
$ eval $(docker-machine env foobarmachine)
$ docker version
Containers: 0
Images: 0
Storage Driver: devicemapper
...
Name: foobarmachine
...
Labels:
 foo=bar
 spam=eggs
 provider=virtualbox
```

The supported flags are as follows:

- `--engine-insecure-registry`: Specify [insecure registries](https://docs.docker.com/reference/commandline/cli/#insecure-registries) to allow with the created engine
- `--engine-registry-mirror`: Specify [registry mirrors](https://github.com/docker/docker/blob/master/docs/sources/articles/registry_mirror.md) to use
- `--engine-label`: Specify [labels](https://docs.docker.com/userguide/labels-custom-metadata/#daemon-labels) for the created engine
- `--engine-storage-driver`: Specify a [storage driver](https://docs.docker.com/reference/commandline/cli/#daemon-storage-driver-option) to use with the engine

If the engine supports specifying the flag multiple times (such as with
`--label`), then so does Docker Machine.

In addition to this subset of daemon flags which are directly supported, Docker
Machine also supports an additional flag, `--engine-opt`, which can be used to
specify arbitrary daemon options with the syntax `--engine-opt flagname=value`.
For example, to specify that the daemon should use `8.8.8.8` as the DNS server
for all containers, and always use the `syslog` [log
driver](https://docs.docker.com/reference/run/#logging-drivers-log-driver) you
could run the following create command:

```
$ docker-machine create -d virtualbox \
    --engine-opt dns=8.8.8.8 \
    --engine-opt log-driver=syslog \
    gdns
```

##### Specifying Docker Swarm options for the created machine

In addition to being able to configure Docker Engine options as listed above,
you can use Machine to specify how the created Swarm master should be
configured). There is a `--swarm-strategy` flag, which you can use to specify
the [scheduling strategy](https://docs.docker.com/swarm/scheduler/strategy/)
which Docker Swarm should use (Machine defaults to the `spread` strategy).
There is also a general purpose `--swarm-opt` option which works similar to how
the aforementioned `--engine-opt` option does, except that it specifies options
for the `swarm manage` command (used to boot a master node) instead of the base
command. You can use this to configure features that power users might be
interested in, such as configuring the heartbeat interval or Swarm's willingness
to over-commit resources.

If you're not sure how to configure these options, it is best to not specify
configuration at all. Docker Machine will choose sensible defaults for you and
you won't have to worry about it.

Example create:

```
$ docker-machine create -d virtualbox \
    --swarm \
    --swarm-master \
    --swarm-discovery token://<token> \
    --swarm-strategy binpack \
    --swarm-opt heartbeat=5 \
    upbeat
```

This will set the swarm scheduling strategy to "binpack" (pack in containers as
tightly as possible per host instead of spreading them out), and the "heartbeat"
interval to 5 seconds.

#### config

Show the Docker client configuration for a machine.

```
$ docker-machine config dev
--tlsverify --tlscacert="/Users/ehazlett/.docker/machines/dev/ca.pem" --tlscert="/Users/ehazlett/.docker/machines/dev/cert.pem" --tlskey="/Users/ehazlett/.docker/machines/dev/key.pem" -H tcp://192.168.99.103:2376
```

#### env

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
# docker-machine.exe env --shell=powershell | Invoke-Expression
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

#### inspect

```
Usage: docker-machine inspect [OPTIONS] [arg...]

Inspect information about a machine

Description:
   Argument is a machine name.

Options:
   --format, -f 	Format the output using the given go template.
```

By default, this will render information about a machine as JSON. If a format is
specified, the given template will be executed for each result.

Go's [text/template](http://golang.org/pkg/text/template/) package
describes all the details of the format.

In addition to the `text/template` syntax, there are some additional functions,
`json` and `prettyjson`, which can be used to format the output as JSON (documented below).

##### Examples

**List all the details of a machine:**

This is the default usage of `inspect`.

```
$ docker-machine inspect dev
{
    "DriverName": "virtualbox",
    "Driver": {
        "MachineName": "docker-host-128be8d287b2028316c0ad5714b90bcfc11f998056f2f790f7c1f43f3d1e6eda",
        "SSHPort": 55834,
        "Memory": 1024,
        "DiskSize": 20000,
        "Boot2DockerURL": "",
        "IPAddress": "192.168.5.99"
    },
    ...
}
```

**Get a machine's IP address:**

For the most part, you can pick out any field from the JSON in a fairly
straightforward manner.

```
$ docker-machine inspect --format='{{.Driver.IPAddress}}' dev
192.168.5.99
```

**Formatting details:**

If you want a subset of information formatted as JSON, you can use the `json`
function in the template.

```
$ docker-machine inspect --format='{{json .Driver}}' dev-fusion
{"Boot2DockerURL":"","CPUS":8,"CPUs":8,"CaCertPath":"/Users/hairyhenderson/.docker/machine/certs/ca.pem","DiskSize":20000,"IPAddress":"172.16.62.129","ISO":"/Users/hairyhenderson/.docker/machine/machines/dev-fusion/boot2docker-1.5.0-GH747.iso","MachineName":"dev-fusion","Memory":1024,"PrivateKeyPath":"/Users/hairyhenderson/.docker/machine/certs/ca-key.pem","SSHPort":22,"SSHUser":"docker","SwarmDiscovery":"","SwarmHost":"tcp://0.0.0.0:3376","SwarmMaster":false}
```

While this is usable, it's not very human-readable. For this reason, there is
`prettyjson`:

```
$ docker-machine inspect --format='{{prettyjson .Driver}}' dev-fusion
{
    "Boot2DockerURL": "",
    "CPUS": 8,
    "CPUs": 8,
    "CaCertPath": "/Users/hairyhenderson/.docker/machine/certs/ca.pem",
    "DiskSize": 20000,
    "IPAddress": "172.16.62.129",
    "ISO": "/Users/hairyhenderson/.docker/machine/machines/dev-fusion/boot2docker-1.5.0-GH747.iso",
    "MachineName": "dev-fusion",
    "Memory": 1024,
    "PrivateKeyPath": "/Users/hairyhenderson/.docker/machine/certs/ca-key.pem",
    "SSHPort": 22,
    "SSHUser": "docker",
    "SwarmDiscovery": "",
    "SwarmHost": "tcp://0.0.0.0:3376",
    "SwarmMaster": false
}
```

#### help

Show help text.

#### ip

Get the IP address of one or more machines.

```
$ docker-machine ip dev
192.168.99.104
$ docker-machine ip dev dev2
192.168.99.104
192.168.99.105
```

#### kill

Kill (abruptly force stop) a machine.

```
$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL
dev    *        virtualbox   Running   tcp://192.168.99.104:2376
$ docker-machine kill dev
$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL
dev    *        virtualbox   Stopped
```

#### ls

```
Usage: docker-machine ls [OPTIONS] [arg...]

List machines

Options:

   --quiet, -q					Enable quiet mode
   --filter [--filter option --filter option]	Filter output based on conditions provided
```

##### Filtering

The filtering flag (`-f` or `--filter)` format is a `key=value` pair. If there is more
than one filter, then pass multiple flags (e.g. `--filter "foo=bar" --filter "bif=baz"`)

The currently supported filters are:

* driver (driver name)
* swarm (swarm master's name)
* state (`Running|Paused|Saved|Stopped|Stopping|Starting|Error`)

##### Examples

```
$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL
dev             virtualbox   Stopped
foo0            virtualbox   Running   tcp://192.168.99.105:2376
foo1            virtualbox   Running   tcp://192.168.99.106:2376
foo2   *        virtualbox   Running   tcp://192.168.99.107:2376
```

```
$ docker-machine ls --filter driver=virtualbox --filter state=Stopped
NAME   ACTIVE   DRIVER       STATE     URL   SWARM
dev             virtualbox   Stopped
```

#### regenerate-certs

Regenerate TLS certificates and update the machine with new certs.

```
$ docker-machine regenerate-certs dev
Regenerate TLS machine certs?  Warning: this is irreversible. (y/n): y
Regenerating TLS certificates
```

#### restart

Restart a machine. Oftentimes this is equivalent to
`docker-machine stop; machine start`.

```
$ docker-machine restart dev
Waiting for VM to start...
```

#### rm

Remove a machine. This will remove the local reference as well as delete it
on the cloud provider or virtualization management platform.

```
$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL
foo0            virtualbox   Running   tcp://192.168.99.105:2376
foo1            virtualbox   Running   tcp://192.168.99.106:2376
$ docker-machine rm foo1
$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL
foo0            virtualbox   Running   tcp://192.168.99.105:2376
```

#### ssh

Log into or run a command on a machine using SSH.

To login, just run `docker-machine ssh machinename`:

```
$ docker-machine ssh dev
                        ##        .
                  ## ## ##       ==
               ## ## ## ##      ===
           /""""""""""""""""\___/ ===
      ~~~ {~~ ~~~~ ~~~ ~~~~ ~~ ~ /  ===- ~~~
           \______ o          __/
             \    \        __/
              \____\______/
 _                 _   ____     _            _
| |__   ___   ___ | |_|___ \ __| | ___   ___| | _____ _ __
| '_ \ / _ \ / _ \| __| __) / _` |/ _ \ / __| |/ / _ \ '__|
| |_) | (_) | (_) | |_ / __/ (_| | (_) | (__|   <  __/ |
|_.__/ \___/ \___/ \__|_____\__,_|\___/ \___|_|\_\___|_|
Boot2Docker version 1.4.0, build master : 69cf398 - Fri Dec 12 01:39:42 UTC 2014
docker@boot2docker:~$ ls /
Users/   dev/     home/    lib/     mnt/     proc/    run/     sys/     usr/
bin/     etc/     init     linuxrc  opt/     root/    sbin/    tmp      var/
```

You can also specify commands to run remotely by appending them directly to the
`docker-machine ssh` command, much like the regular `ssh` program works:

```
$ docker-machine ssh dev free
             total         used         free       shared      buffers
Mem:       1023556       183136       840420            0        30920
-/+ buffers:             152216       871340
Swap:      1212036            0      1212036
```

If the command you are appending has flags, e.g. `df -h`, you can use the flag
parsing terminator (`--`) to avoid confusing the `docker-machine` client, which
will otherwise interpret them as flags you intended to pass to it:

```
$ docker-machine ssh dev -- df -h
Filesystem                Size      Used Available Use% Mounted on
rootfs                  899.6M     85.9M    813.7M  10% /
tmpfs                   899.6M     85.9M    813.7M  10% /
tmpfs                   499.8M         0    499.8M   0% /dev/shm
/dev/sda1                18.2G     58.2M     17.2G   0% /mnt/sda1
cgroup                  499.8M         0    499.8M   0% /sys/fs/cgroup
/dev/sda1                18.2G     58.2M     17.2G   0%
/mnt/sda1/var/lib/docker/aufs
```

##### Different types of SSH

When Docker Machine is invoked, it will check to see if you have the venerable
`ssh` binary around locally and will attempt to use that for the SSH commands it
needs to run, whether they are a part of an operation such as creation or have
been requested by the user directly. If it does not find an external `ssh`
binary locally, it will default to using a native Go implementation from
[crypto/ssh](https://godoc.org/golang.org/x/crypto/ssh). This is useful in
situations where you may not have access to traditional UNIX tools, such as if
you are using Docker Machine on Windows without having msysgit installed
alongside of it.

In most situations, you will not have to worry about this implementation detail
and Docker Machine will act sensibly out of the box. However, if you
deliberately want to use the Go native version, you can do so with a global
command line flag / environment variable like so:

```
$ docker-machine --native-ssh ssh dev
```

There are some variations in behavior between the two methods, so please report
any issues or inconsistencies if you come across them.

#### scp

Copy files from your local host to a machine, from machine to machine, or from a
machine to your local host using `scp`.

The notation is `machinename:/path/to/files` for the arguments; in the host
machine's case, you don't have to specify the name, just the path.

Consider the following example:

```
$ cat foo.txt
cat: foo.txt: No such file or directory
$ docker-machine ssh dev pwd
/home/docker
$ docker-machine ssh dev 'echo A file created remotely! >foo.txt'
$ docker-machine scp dev:/home/docker/foo.txt .
foo.txt                                                           100%   28     0.0KB/s   00:00
$ cat foo.txt
A file created remotely!
```

Files are copied recursively by default (`scp`'s `-r` flag).

In the case of transferring files from machine to machine, they go through the
local host's filesystem first (using `scp`'s `-3` flag).

#### start

Start a machine.

```
$ docker-machine start dev
Starting VM...
```

#### stop

Gracefully stop a machine.

```
$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL
dev    *        virtualbox   Running   tcp://192.168.99.104:2376
$ docker-machine stop dev
$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL
dev    *        virtualbox   Stopped
```

#### upgrade

Upgrade a machine to the latest version of Docker. If the machine uses Ubuntu
as the underlying operating system, it will upgrade the package `lxc-docker`
(our recommended install method). If the machine uses boot2docker, this command
will download the latest boot2docker ISO and replace the machine's existing ISO
with the latest.

```
$ docker-machine upgrade dev
Stopping machine to do the upgrade...
Upgrading machine dev...
Downloading latest boot2docker release to /home/username/.docker/machine/cache/boot2docker.iso...
Starting machine back up...
Waiting for VM to start...
```

> **Note**: If you are using a custom boot2docker ISO specified using
> `--virtualbox-boot2docker-url` or an equivalent flag, running an upgrade on
> that machine will completely replace the specified ISO with the latest
> "vanilla" boot2docker ISO available.

#### url

Get the URL of a host

```
$ docker-machine url dev
tcp://192.168.99.109:2376
```

## Drivers

#### Amazon Web Services
Create machines on [Amazon Web Services](http://aws.amazon.com). You will need an Access Key ID, Secret Access Key and a VPC ID. To find the VPC ID, login to the AWS console and go to Services -> VPC -> Your VPCs. Select the one where you would like to launch the instance.

Options:

 - `--amazonec2-access-key`: **required** Your access key id for the Amazon Web Services API.
 - `--amazonec2-secret-key`: **required** Your secret access key for the Amazon Web Services API.
 - `--amazonec2-session-token`: Your session token for the Amazon Web Services API.
 - `--amazonec2-ami`: The AMI ID of the instance to use.
 - `--amazonec2-region`: The region to use when launching the instance.
 - `--amazonec2-vpc-id`: **required** Your VPC ID to launch the instance in.
 - `--amazonec2-zone`: The AWS zone to launch the instance in (i.e. one of a,b,c,d,e).
 - `--amazonec2-subnet-id`: AWS VPC subnet id.
 - `--amazonec2-security-group`: AWS VPC security group name.
 - `--amazonec2-instance-type`: The instance type to run.
 - `--amazonec2-root-size`: The root disk size of the instance (in GB).
 - `--amazonec2-iam-instance-profile`: The AWS IAM role name to be used as the instance profile.
 - `--amazonec2-ssh-user`: SSH Login user name.
 - `--amazonec2-request-spot-instance`: Use spot instances.
 - `--amazonec2-spot-price`: Spot instance bid price (in dollars). Require the `--amazonec2-request-spot-instance` flag.
 - `--amazonec2-private-address-only`: Use the private IP address only.
 - `--amazonec2-monitoring`: Enable CloudWatch Monitoring.

By default, the Amazon EC2 driver will use a daily image of Ubuntu 14.04 LTS.

| Region         | AMI ID       |
|----------------|--------------|
| ap-northeast-1 | ami-f4b06cf4 |
| ap-southeast-1 | ami-b899a2ea |
| ap-southeast-2 | ami-b59ce48f |
| cn-north-1     | ami-da930ee3 |
| eu-west-1      | ami-45d8a532 |
| eu-central-1   | ami-b6e0d9ab |
| sa-east-1      | ami-1199190c |
| us-east-1      | ami-5f709f34 |
| us-west-1      | ami-615cb725 |
| us-west-2      | ami-7f675e4f |
| us-gov-west-1  | ami-99a9c9ba |

Environment variables and default values:

| CLI option                          | Environment variable    | Default          |
|-------------------------------------|-------------------------|------------------|
| **`--amazonec2-access-key`**        | `AWS_ACCESS_KEY_ID`     | -                |
| **`--amazonec2-secret-key`**        | `AWS_SECRET_ACCESS_KEY` | -                |
| `--amazonec2-session-token`         | `AWS_SESSION_TOKEN`     | -                |
| `--amazonec2-ami`                   | `AWS_AMI`               | `ami-5f709f34`   |
| `--amazonec2-region`                | `AWS_DEFAULT_REGION`    | `us-east-1`      |
| **`--amazonec2-vpc-id`**            | `AWS_VPC_ID`            | -                |
| `--amazonec2-zone`                  | `AWS_ZONE`              | `a`              |
| `--amazonec2-subnet-id`             | `AWS_SUBNET_ID`         | -                |
| `--amazonec2-security-group`        | `AWS_SECURITY_GROUP`    | `docker-machine` |
| `--amazonec2-instance-type`         | `AWS_INSTANCE_TYPE`     | `t2.micro`       |
| `--amazonec2-root-size`             | `AWS_ROOT_SIZE`         | `16`             |
| `--amazonec2-iam-instance-profile`  | `AWS_INSTANCE_PROFILE`  | -                |
| `--amazonec2-ssh-user`              | `AWS_SSH_USER`          | `ubuntu`         |
| `--amazonec2-request-spot-instance` | -                       | `false`          |
| `--amazonec2-spot-price`            | -                       | `0.50`           |
| `--amazonec2-private-address-only`  | -                       | `false`          |
| `--amazonec2-monitoring`            | -                       | `false`          |

#### Digital Ocean
Create Docker machines on [Digital Ocean](https://www.digitalocean.com/).

You need to create a personal access token under "Apps & API" in the Digital Ocean
Control Panel and pass that to `docker-machine create` with the `--digitalocean-access-token` option.

    $ docker-machine create --driver digitalocean --digitalocean-access-token=aa9399a2175a93b17b1c86c807e08d3fc4b79876545432a629602f61cf6ccd6b test-this

Options:

 - `--digitalocean-access-token`: **required** Your personal access token for the Digital Ocean API.
 - `--digitalocean-image`: The name of the Digital Ocean image to use.
 - `--digitalocean-region`: The region to create the droplet in, see [Regions API](https://developers.digitalocean.com/documentation/v2/#regions) for how to get a list.
 - `--digitalocean-size`: The size of the Digital Ocean droplet (larger than default options are of the form `2gb`).
 - `--digitalocean-ipv6`: Enable IPv6 support for the droplet.
 - `--digitalocean-private-networking`: Enable private networking support for the droplet.
 - `--digitalocean-backups`: Enable Digital Oceans backups for the droplet.

The DigitalOcean driver will use `ubuntu-14-04-x64` as the default image.

Environment variables and default values:

| CLI option                          | Environment variable              | Default  |
|-------------------------------------|-----------------------------------|----------|
| **`--digitalocean-access-token`**   | `DIGITALOCEAN_ACCESS_TOKEN`       | -        |
| `--digitalocean-image`              | `DIGITALOCEAN_IMAGE`              | `docker` |
| `--digitalocean-region`             | `DIGITALOCEAN_REGION`             | `nyc3`   |
| `--digitalocean-size`               | `DIGITALOCEAN_SIZE`               | `512mb`  |
| `--digitalocean-ipv6`               | `DIGITALOCEAN_IPV6`               | `false`  |
| `--digitalocean-private-networking` | `DIGITALOCEAN_PRIVATE_NETWORKING` | `false`  |
| `--digitalocean-backups`            | `DIGITALOCEAN_BACKUPS`            | `false`  |

#### exoscale
Create machines on [exoscale](https://www.exoscale.ch/).

Get your API key and API secret key from [API details](https://portal.exoscale.ch/account/api) and pass them to `machine create` with the `--exoscale-api-key` and `--exoscale-api-secret-key` options.

Options:

 - `--exoscale-url`: Your API endpoint.
 - `--exoscale-api-key`: **required** Your API key.
 - `--exoscale-api-secret-key`: **required** Your API secret key.
 - `--exoscale-instance-profile`: Instance profile.
 - `--exoscale-disk-size`: Disk size for the host in GB.
 - `--exoscale-image`: exoscale disk size. (10, 50, 100, 200, 400)
 - `--exoscale-security-group`: Security group. It will be created if it doesn't exist.
 - `--exoscale-availability-zone`: exoscale availability zone.

If a custom security group is provided, you need to ensure that you allow TCP ports 22 and 2376 in an ingress rule. Moreover, if you want to use Swarm, also add TCP port 3376.

Environment variables and default values:

| CLI option                      | Environment variable         | Default                           |
|---------------------------------|------------------------------|-----------------------------------|
| `--exoscale-url`                | `EXOSCALE_ENDPOINT`          | `https://api.exoscale.ch/compute` |
| **`--exoscale-api-key`**        | `EXOSCALE_API_KEY`           | -                                 |
| **`--exoscale-api-secret-key`** | `EXOSCALE_API_SECRET`        | -                                 |
| `--exoscale-instance-profile`   | `EXOSCALE_INSTANCE_PROFILE`  | `small`                           |
| `--exoscale-disk-size`          | `EXOSCALE_DISK_SIZE`         | `50`                              |
| `--exoscale-image`              | `EXSOCALE_IMAGE`             | `ubuntu-14.04`                    |
| `--exoscale-security-group`     | `EXOSCALE_SECURITY_GROUP`    | `docker-machine`                  |
| `--exoscale-availability-zone`  | `EXOSCALE_AVAILABILITY_ZONE` | `ch-gva-2`                        |

#### Generic
Create machines using an existing VM/Host with SSH.

This is useful if you are using a provider that Machine does not support
directly or if you would like to import an existing host to allow Docker
Machine to manage.

Options:

 - `--generic-ip-address`: **required** IP Address of host.
 - `--generic-ssh-user`: SSH username used to connect.
 - `--generic-ssh-key`: Path to the SSH user private key.
 - `--generic-ssh-port`: Port to use for SSH.

> **Note**: You must use a base operating system supported by Machine.

Environment variables and default values:

| CLI option                 | Environment variable | Default             |
|----------------------------|----------------------|---------------------|
| **`--generic-ip-address`** | -                    | -                   |
| `--generic-ssh-user`       | -                    | `root`              |
| `--generic-ssh-key`        | -                    | `$HOME/.ssh/id_rsa` |
| `--generic-ssh-port`       | -                    | `22`                |

#### Google Compute Engine
Create machines on [Google Compute Engine](https://cloud.google.com/compute/). You will need a Google account and project name. See https://cloud.google.com/compute/docs/projects for details on projects.

The Google driver uses oAuth. When creating the machine, you will have your browser opened to authorize. Once authorized, paste the code given in the prompt to launch the instance.

Options:

 - `--google-zone`: The zone to launch the instance.
 - `--google-machine-type`: The type of instance.
 - `--google-username`: The username to use for the instance.
 - `--google-project`: **required** The name of your project to use when launching the instance.
 - `--google-auth-token`: Your oAuth token for the Google Cloud API.
 - `--google-scopes`: The scopes for OAuth 2.0 to Access Google APIs. See [Google Compute Engine Doc](https://cloud.google.com/storage/docs/authentication).
 - `--google-disk-size`: The disk size of instance.
 - `--google-disk-type`: The disk type of instance.

The GCE driver will use the `ubuntu-1404-trusty-v20150316` instance type unless otherwise specified.

Environment variables and default values:

| CLI option                | Environment variable  | Default                              |
|---------------------------|-----------------------|--------------------------------------|
| `--google-zone`           | `GOOGLE_ZONE`         | `us-central1-a`                      |
| `--google-machine-type`   | `GOOGLE_MACHINE_TYPE` | `f1-micro`                           |
| `--google-username`       | `GOOGLE_USERNAME`     | `docker-user`                        |
| **`--google-project`**    | `GOOGLE_PROJECT`      | -                                    |
| `--google-auth-token`     | `GOOGLE_AUTH_TOKEN`   | -                                    |
| `--google-scopes`         | `GOOGLE_SCOPES`       | `devstorage.read_only,logging.write` |
| `--google-disk-size`      | `GOOGLE_DISK_SIZE`    | `10`                                 |
| `--google-disk-type`      | `GOOGLE_DISK_TYPE`    | `pd-standard`                        |

#### IBM Softlayer
Create machines on [Softlayer](http://softlayer.com).

You need to generate an API key in the softlayer control panel.
[Retrieve your API key](http://knowledgelayer.softlayer.com/procedure/retrieve-your-api-key)

Options:

  - `--softlayer-memory`: Memory for host in MB.
  - `--softlayer-disk-size`: A value of `0` will set the SoftLayer default.
  - `--softlayer-user`: **required** Username for your SoftLayer account, api key needs to match this user.
  - `--softlayer-api-key`: **required** API key for your user account.
  - `--softlayer-region`: SoftLayer region.
  - `--softlayer-cpu`: Number of CPUs for the machine.
  - `--softlayer-hostname`: Hostname for the machine.
  - `--softlayer-domain`: **required** Domain name for the machine.
  - `--softlayer-api-endpoint`: Change SoftLayer API endpoint.
  - `--softlayer-hourly-billing`: Specifies that hourly billing should be used (default), otherwise monthly billing is used.
  - `--softlayer-local-disk`: Use local machine disk instead of SoftLayer SAN.
  - `--softlayer-private-net-only`: Disable public networking.
  - `--softlayer-image`: OS Image to use.
  - `--softlayer-public-vlan-id`: Your public VLAN ID.
  - `--softlayer-private-vlan-id`: Your private VLAN ID.

The SoftLayer driver will use `UBUNTU_LATEST` as the image type by default.

Environment variables and default values:

| CLI option                     | Environment variable        | Default                     |
|--------------------------------|-----------------------------|-----------------------------|
| `--softlayer-memory`           | `SOFTLAYER_MEMORY`          | `1024`                      |
| `--softlayer-disk-size`        | `SOFTLAYER_DISK_SIZE`       | `0`                         |
| **`--softlayer-user`**         | `SOFTLAYER_USER`            | -                           |
| **`--softlayer-api-key`**      | `SOFTLAYER_API_KEY`         | -                           |
| `--softlayer-region`           | `SOFTLAYER_REGION`          | `dal01`                     |
| `--softlayer-cpu`              | `SOFTLAYER_CPU`             | `1`                         |
| `--softlayer-hostname`         | `SOFTLAYER_HOSTNAME`        | `docker`                    |
| **`--softlayer-domain`**       | `SOFTLAYER_DOMAIN`          | -                           |
| `--softlayer-api-endpoint`     | `SOFTLAYER_API_ENDPOINT`    | `api.softlayer.com/rest/v3` |
| `--softlayer-hourly-billing`   | `SOFTLAYER_HOURLY_BILLING`  | `false`                     |
| `--softlayer-local-disk`       | `SOFTLAYER_LOCAL_DISK`      | `false`                     |
| `--softlayer-private-net-only` | `SOFTLAYER_PRIVATE_NET`     | `false`                     |
| `--softlayer-image`            | `SOFTLAYER_IMAGE`           | `UBUNTU_LATEST`             |
| `--softlayer-public-vlan-id`   | `SOFTLAYER_PUBLIC_VLAN_ID`  | `0`                         |
| `--softlayer-private-vlan-id`  | `SOFTLAYER_PRIVATE_VLAN_ID` | `0`                         |

#### Microsoft Azure
Create machines on [Microsoft Azure](http://azure.microsoft.com/).

You need to create a subscription with a cert. Run these commands and answer the questions:

    $ openssl req -x509 -nodes -days 365 -newkey rsa:1024 -keyout mycert.pem -out mycert.pem
    $ openssl pkcs12 -export -out mycert.pfx -in mycert.pem -name "My Certificate"
    $ openssl x509 -inform pem -in mycert.pem -outform der -out mycert.cer

Go to the Azure portal, go to the "Settings" page (you can find the link at the bottom of the
left sidebar - you need to scroll), then "Management Certificates" and upload `mycert.cer`.

Grab your subscription ID from the portal, then run `docker-machine create` with these details:

    $ docker-machine create -d azure --azure-subscription-id="SUB_ID" --azure-subscription-cert="mycert.pem" A-VERY-UNIQUE-NAME

The Azure driver uses the `b39f27a8b8c64d52b05eac6a62ebad85__Ubuntu-14_04_1-LTS-amd64-server-20140927-en-us-30GB`
image by default. Note, this image is not available in the Chinese regions. In China you should
 specify `b549f4301d0b4295b8e76ceb65df47d4__Ubuntu-14_04_1-LTS-amd64-server-20140927-en-us-30GB`.

You may need to `machine ssh` in to the virtual machine and reboot to ensure that the OS is updated.

Options:

 - `--azure-docker-port`: Port for Docker daemon.
 - `--azure-image`: Azure image name. See [How to: Get the Windows Azure Image Name](https://msdn.microsoft.com/en-us/library/dn135249%28v=nav.70%29.aspx)
 - `--azure-location`: Machine instance location.
 - `--azure-password`: Your Azure password.
 - `--azure-publish-settings-file`: Azure setting file. See [How to: Download and Import Publish Settings and Subscription Information](https://msdn.microsoft.com/en-us/library/dn385850%28v=nav.70%29.aspx)
 - `--azure-size`: Azure disk size.
 - `--azure-ssh-port`: Azure SSH port.
 - `--azure-subscription-id`: **required** Your Azure subscription ID (A GUID like `d255d8d7-5af0-4f5c-8a3e-1545044b861e`).
 - `--azure-subscription-cert`: **required** Your Azure subscription cert.
 - `--azure-username`: Azure login user name.

Environment variables and default values:

| CLI option                      | Environment variable          | Default               |
|---------------------------------|-------------------------------| ----------------------|
| `--azure-docker-port`           | -                             | `2376`                |
| `--azure-image`                 | `AZURE_IMAGE`                 | *Ubuntu 14.04 LTS x64*|
| `--azure-location`              | `AZURE_LOCATION`              | `West US`             |
| `--azure-password`              | -                             | -                     |
| `--azure-publish-settings-file` | `AZURE_PUBLISH_SETTINGS_FILE` | -                     |
| `--azure-size`                  | `AZURE_SIZE`                  | `Small`               |
| `--azure-ssh-port`              | -                             | `22`                  |
| **`--azure-subscription-cert`** | `AZURE_SUBSCRIPTION_CERT`     | -                     |
| **`--azure-subscription-id`**   | `AZURE_SUBSCRIPTION_ID`       | -                     |
| `--azure-username`              | -                             | `ubuntu`              |

#### Microsoft Hyper-V
Creates a Boot2Docker virtual machine locally on your Windows machine
using Hyper-V. [See here](http://windows.microsoft.com/en-us/windows-8/hyper-v-run-virtual-machines)
for instructions to enable Hyper-V. You will need to use an
Administrator level account to create and manage Hyper-V machines.

> **Note**: You will need an existing virtual switch to use the
> driver. Hyper-V can share an external network interface (aka
> bridging), see [this blog](http://blogs.technet.com/b/canitpro/archive/2014/03/11/step-by-step-enabling-hyper-v-for-use-on-windows-8-1.aspx).
> If you would like to use NAT, create an internal network, and use
> [Internet Connection
> Sharing](http://www.packet6.com/allowing-windows-8-1-hyper-v-vm-to-work-with-wifi/).

Options:

 - `--hyper-v-boot2docker-url`: The URL of the boot2docker ISO. Defaults to the latest available version.
 - `--hyper-v-boot2docker-location`: Location of a local boot2docker iso to use. Overrides the URL option below.
 - `--hyper-v-virtual-switch`: Name of the virtual switch to use. Defaults to first found.
 - `--hyper-v-disk-size`: Size of disk for the host in MB.
 - `--hyper-v-memory`: Size of memory for the host in MB. By default, the machine is setup to use dynamic memory.

Environment variables and default values:

| CLI option                       | Environment variable | Default                  |
|----------------------------------|----------------------| -------------------------|
| `--hyper-v-boot2docker-url`      | -                    | *Latest boot2docker url* |
| `--hyper-v-boot2docker-location` | -                    | -                        |
| `--hyper-v-virtual-switch`       | -                    | *first found*            |
| `--hyper-v-disk-size`            | -                    | `20000`                  |
| `--hyper-v-memory`               | -                    | `1024`                   |

#### OpenStack
Create machines on [OpenStack](http://www.openstack.org/software/)

Mandatory:

 - `--openstack-auth-url`: Keystone service base URL.
 - `--openstack-flavor-id` or `--openstack-flavor-name`: Identify the flavor that will be used for the machine.
 - `--openstack-image-id` or `--openstack-image-name`: Identify the image that will be used for the machine.

Options:

 - `--openstack-insecure`: Explicitly allow openstack driver to perform "insecure" SSL (https) requests. The server's certificate will not be verified against any certificate authorities. This option should be used with caution.
 - `--openstack-domain-name` or `--openstack-domain-id`: Domain to use for authentication (Keystone v3 only)
 - `--openstack-username`: User identifier to authenticate with.
 - `--openstack-password`: User password. It can be omitted if the standard environment variable `OS_PASSWORD` is set.
 - `--openstack-tenant-name` or `--openstack-tenant-id`: Identify the tenant in which the machine will be created.
 - `--openstack-region`: The region to work on. Can be omitted if there is only one region on the OpenStack.
 - `--openstack-availability-zone`: The availability zone in which to launch the server.
 - `--openstack-endpoint-type`: Endpoint type can be `internalURL`, `adminURL` on `publicURL`. If is a helper for the driver
   to choose the right URL in the OpenStack service catalog. If not provided the default id `publicURL`
 - `--openstack-net-name` or `--openstack-net-id`: Identify the private network the machine will be connected on. If your OpenStack project project contains only one private network it will be use automatically.
 - `--openstack-sec-groups`: If security groups are available on your OpenStack you can specify a comma separated list
   to use for the machine (e.g. `secgrp001,secgrp002`).
 - `--openstack-floatingip-pool`: The IP pool that will be used to get a public IP can assign it to the machine. If there is an
   IP address already allocated but not assigned to any machine, this IP will be chosen and assigned to the machine. If
   there is no IP address already allocated a new IP will be allocated and assigned to the machine.
 - `--openstack-ssh-user`: The username to use for SSH into the machine. If not provided `root` will be used.
 - `--openstack-ssh-port`: Customize the SSH port if the SSH server on the machine does not listen on the default port.

Environment variables and default values:

| CLI option                       | Environment variable   | Default |
|----------------------------------|------------------------|---------|
| `--openstack-auth-url`           | `OS_AUTH_URL`          | -       |
| `--openstack-flavor-name`        | -                      | -       |
| `--openstack-flavor-id`          | -                      | -       |
| `--openstack-image-name`         | -                      | -       |
| `--openstack-image-id`           | -                      | -       |
| `--openstack-insecure`           | -                      | -       |
| `--openstack-domain-name`        | `OS_DOMAIN_NAME`       | -       |
| `--openstack-domain-id`          | `OS_DOMAIN_ID`         | -       |
| `--openstack-username`           | `OS_USERNAME`          | -       |
| `--openstack-password`           | `OS_PASSWORD`          | -       |
| `--openstack-tenant-name`        | `OS_TENANT_NAME`       | -       |
| `--openstack-tenant-id`          | `OS_TENANT_ID`         | -       |
| `--openstack-region`             | `OS_REGION_NAME`       | -       |
| `--openstack-availability-zone`  | `OS_AVAILABILITY_ZONE` | -       |
| `--openstack-endpoint-type`      | `OS_ENDPOINT_TYPE`     | -       |
| `--openstack-net-name`           | -                      | -       |
| `--openstack-net-id`             | -                      | -       |
| `--openstack-sec-groups`         | -                      | -       |
| `--openstack-floatingip-pool`    | -                      | -       |
| `--openstack-ssh-user`           | -                      | `root`  |
| `--openstack-ssh-port`           | -                      | `22`    |

#### Rackspace
Create machines on [Rackspace cloud](http://www.rackspace.com/cloud)

Options:

 - `--rackspace-username`: **required** Rackspace account username.
 - `--rackspace-api-key`: **required** Rackspace API key.
 - `--rackspace-region`: **required** Rackspace region name.
 - `--rackspace-endpoint-type`: Rackspace endpoint type (`adminURL`, `internalURL` or the default `publicURL`).
 - `--rackspace-image-id`: Rackspace image ID. Default: Ubuntu 14.10 (Utopic Unicorn) (PVHVM).
 - `--rackspace-flavor-id`: Rackspace flavor ID. Default: General Purpose 1GB.
 - `--rackspace-ssh-user`: SSH user for the newly booted machine.
 - `--rackspace-ssh-port`: SSH port for the newly booted machine.
 - `--rackspace-docker-install`: Set if Docker has to be installed on the machine.

The Rackspace driver will use `598a4282-f14b-4e50-af4c-b3e52749d9f9` (Ubuntu 14.04 LTS) by default.

Environment variables and default values:

| CLI option                   | Environment variable | Default                                |
|------------------------------|----------------------|----------------------------------------|
| **`--rackspace-username`**   | `OS_USERNAME`        | -                                      |
| **`--rackspace-api-key`**    | `OS_API_KEY`         | -                                      |
| **`--rackspace-region`**     | `OS_REGION_NAME`     | -                                      |
| `--rackspace-endpoint-type`  | `OS_ENDPOINT_TYPE`   | `publicURL`                            |
| `--rackspace-image-id`       | -                    | `598a4282-f14b-4e50-af4c-b3e52749d9f9` |
| `--rackspace-flavor-id`      | `OS_FLAVOR_ID`       | `general1-1`                           |
| `--rackspace-ssh-user`       | -                    | `root`                                 |
| `--rackspace-ssh-port`       | -                    | `22`                                   |
| `--rackspace-docker-install` | -                    | `true`                                 |

#### Oracle VirtualBox
Create machines locally using [VirtualBox](https://www.virtualbox.org/).
This driver requires VirtualBox to be installed on your host.

    $ docker-machine create --driver=virtualbox vbox-test
    
You can create an entirely new machine or you can convert a Boot2Docker VM into
a machine by importing the VM. To convert a Boot2Docker VM, you'd use the following
command:

    $ docker-machine create -d virtualbox --virtualbox-import-boot2docker-vm boot2docker-vm b2d


Options:

 - `--virtualbox-memory`: Size of memory for the host in MB.
 - `--virtualbox-cpu-count`: Number of CPUs to use to create the VM. Defaults to single CPU.
 - `--virtualbox-disk-size`: Size of disk for the host in MB.
 - `--virtualbox-boot2docker-url`: The URL of the boot2docker image. Defaults to the latest available version.
 - `--virtualbox-import-boot2docker-vm`: The name of a Boot2Docker VM to import.
 - `--virtualbox-hostonly-cidr`: The CIDR of the host only adapter.

The `--virtualbox-boot2docker-url` flag takes a few different forms. By
default, if no value is specified for this flag, Machine will check locally for
a boot2docker ISO. If one is found, that will be used as the ISO for the
created machine. If one is not found, the latest ISO release available on
[boot2docker/boot2docker](https://github.com/boot2docker/boot2docker) will be
downloaded and stored locally for future use. Note that this means you must run
`docker-machine upgrade` deliberately on a machine if you wish to update the "cached"
boot2docker ISO.

This is the default behavior (when `--virtualbox-boot2docker-url=""`), but the
option also supports specifying ISOs by the `http://` and `file://` protocols.
`file://` will look at the path specified locally to locate the ISO: for
instance, you could specify `--virtualbox-boot2docker-url
file://$HOME/Downloads/rc.iso` to test out a release candidate ISO that you have
downloaded already. You could also just get an ISO straight from the Internet
using the `http://` form.

To customize the host only adapter, you can use the `--virtualbox-hostonly-cidr`
flag.  This will specify the host IP and Machine will calculate the VirtualBox
DHCP server address (a random IP on the subnet between `.1` and `.25`) so 
it does not clash with the specified host IP.
Machine will also specify the DHCP lower bound to `.100` and the upper bound
to `.254`.  For example, a specified CIDR of `192.168.24.1/24` would have a
DHCP server between `192.168.24.2-25`, a lower bound of `192.168.24.100` and 
upper bound of `192.168.24.254`.

Environment variables and default values:

| CLI option                           | Environment variable         | Default                  |
|--------------------------------------|------------------------------|--------------------------|
| `--virtualbox-memory`                | `VIRTUALBOX_MEMORY_SIZE`     | `1024`                   |
| `--virtualbox-cpu-count`             | `VIRTUALBOX_CPU_COUNT`       | `1`                      |
| `--virtualbox-disk-size`             | `VIRTUALBOX_DISK_SIZE`       | `20000`                  |
| `--virtualbox-boot2docker-url`       | `VIRTUALBOX_BOOT2DOCKER_URL` | *Latest boot2docker url* |
| `--virtualbox-import-boot2docker-vm` | -                            | `boot2docker-vm`         |
| `--virtualbox-hostonly-cidr`         | `VIRTUALBOX_HOSTONLY_CIDR`   | `192.168.99.1/24`        |

#### VMware Fusion
Creates machines locally on [VMware Fusion](http://www.vmware.com/products/fusion). Requires VMware Fusion to be installed.

Options:

 - `--vmwarefusion-boot2docker-url`: URL for boot2docker image.
 - `--vmwarefusion-cpu-count`: Number of CPUs for the machine (-1 to use the number of CPUs available)
 - `--vmwarefusion-disk-size`: Size of disk for host VM (in MB).
 - `--vmwarefusion-memory-size`: Size of memory for host VM (in MB).

The VMware Fusion driver uses the latest boot2docker image. 
See [frapposelli/boot2docker](https://github.com/frapposelli/boot2docker/tree/vmware-64bit)

Environment variables and default values:

| CLI option                       | Environment variable     | Default                  |
|----------------------------------|--------------------------|--------------------------|
| `--vmwarefusion-boot2docker-url` | `FUSION_BOOT2DOCKER_URL` | *Latest boot2docker url* |
| `--vmwarefusion-cpu-count`       | `FUSION_CPU_COUNT`       | `1`                      |
| `--vmwarefusion-disk-size`       | `FUSION_MEMORY_SIZE`     | `20000`                  |
| `--vmwarefusion-memory-size`     | `FUSION_DISK_SIZE`       | `1024`                   |

#### VMware vCloud Air
Creates machines on [vCloud Air](http://vcloud.vmware.com) subscription service. You need an account within an existing subscription of vCloud Air VPC or Dedicated Cloud.

Options:

 - `--vmwarevcloudair-username`: **required** vCloud Air Username.
 - `--vmwarevcloudair-password`: **required** vCloud Air Password.
 - `--vmwarevcloudair-computeid`: Compute ID (if using Dedicated Cloud).
 - `--vmwarevcloudair-vdcid`: Virtual Data Center ID.
 - `--vmwarevcloudair-orgvdcnetwork`: Organization VDC Network to attach.
 - `--vmwarevcloudair-edgegateway`: Organization Edge Gateway.
 - `--vmwarevcloudair-publicip`: Org Public IP to use.
 - `--vmwarevcloudair-catalog`: Catalog.
 - `--vmwarevcloudair-catalogitem`: Catalog Item.
 - `--vmwarevcloudair-provision`: Install Docker binaries.
 - `--vmwarevcloudair-cpu-count`: VM CPU Count.
 - `--vmwarevcloudair-memory-size`: VM Memory Size in MB.
 - `--vmwarevcloudair-ssh-port`: SSH port.
 - `--vmwarevcloudair-docker-port`: Docker port.

The VMware vCloud Air driver will use the `Ubuntu Server 12.04 LTS (amd64 20140927)` image by default.

Environment variables and default values:

| CLI option                        | Environment variable      | Default                                    |
|-----------------------------------|---------------------------|--------------------------------------------|
| **`--vmwarevcloudair-username`**  | `VCLOUDAIR_USERNAME`      | -                                          |
| **`--vmwarevcloudair-password`**  | `VCLOUDAIR_PASSWORD`      | -                                          |
| `--vmwarevcloudair-computeid`     | `VCLOUDAIR_COMPUTEID`     | -                                          |
| `--vmwarevcloudair-vdcid`         | `VCLOUDAIR_VDCID`         | -                                          |
| `--vmwarevcloudair-orgvdcnetwork` | `VCLOUDAIR_ORGVDCNETWORK` | `<vdcid>-default-routed`                   |
| `--vmwarevcloudair-edgegateway`   | `VCLOUDAIR_EDGEGATEWAY`   | `<vdcid>`                                  |
| `--vmwarevcloudair-publicip`      | `VCLOUDAIR_PUBLICIP`      | -                                          |
| `--vmwarevcloudair-catalog`       | `VCLOUDAIR_CATALOG`       | `Public Catalog`                           |
| `--vmwarevcloudair-catalogitem`   | `VCLOUDAIR_CATALOGITEM`   | `Ubuntu Server 12.04 LTS (amd64 20140927)` |
| `--vmwarevcloudair-provision`     | `VCLOUDAIR_PROVISION`     | `true`                                     |
| `--vmwarevcloudair-cpu-count`     | `VCLOUDAIR_CPU_COUNT`     | `1`                                        |
| `--vmwarevcloudair-memory-size`   | `VCLOUDAIR_MEMORY_SIZE`   | `2048`                                     |
| `--vmwarevcloudair-ssh-port`      | `VCLOUDAIR_SSH_PORT`      | `22`                                       |
| `--vmwarevcloudair-docker-port`   | `VCLOUDAIR_DOCKER_PORT`   | `2376`                                     |

#### VMware vSphere
Creates machines on a [VMware vSphere](http://www.vmware.com/products/vsphere) Virtual Infrastructure. Requires a working vSphere (ESXi with non free license or 60 days trial and optionally vCenter) installation. The vSphere driver depends on [`govc`](https://github.com/vmware/govmomi/tree/master/govc) (must be in path) and has been tested with [vmware/govmomi@`c848630`](https://github.com/vmware/govmomi/commit/c8486300bfe19427e4f3226e3b3eac067717ef17).

Options:

 - `--vmwarevsphere-cpu-count`: CPU number for Docker VM.
 - `--vmwarevsphere-memory-size`: Size of memory for Docker VM (in MB).
 - `--vmwarevsphere-boot2docker-url`: URL for boot2docker image.
 - `--vmwarevsphere-vcenter`: IP/hostname for vCenter (or ESXi if connecting directly to a single host).
 - `--vmwarevsphere-disk-size`: Size of disk for Docker VM (in MB).
 - `--vmwarevsphere-username`: **required** vSphere Username.
 - `--vmwarevsphere-password`: **required** vSphere Password.
 - `--vmwarevsphere-network`: Network where the Docker VM will be attached.
 - `--vmwarevsphere-datastore`: Datastore for Docker VM.
 - `--vmwarevsphere-datacenter`: Datacenter for Docker VM (must be set to `ha-datacenter` when connecting to a single host).
 - `--vmwarevsphere-pool`: Resource pool for Docker VM.
 - `--vmwarevsphere-compute-ip`: Compute host IP where the Docker VM will be instantiated.

The VMware vSphere driver uses the latest boot2docker image.

Environment variables and default values:

| CLI option                        | Environment variable      | Default                  |
|-----------------------------------|---------------------------|--------------------------|
| `--vmwarevsphere-cpu-count`       | `VSPHERE_CPU_COUNT`       | `2`                      |
| `--vmwarevsphere-memory-size`     | `VSPHERE_MEMORY_SIZE`     | `2048`                   |
| `--vmwarevsphere-disk-size`       | `VSPHERE_DISK_SIZE`       | `20000`                  |
| `--vmwarevsphere-boot2docker-url` | `VSPHERE_BOOT2DOCKER_URL` | *Latest boot2docker url* |
| `--vmwarevsphere-vcenter`         | `VSPHERE_VCENTER`         | -                        |
| **`--vmwarevsphere-username`**    | `VSPHERE_USERNAME`        | -                        |
| **`--vmwarevsphere-password`**    | `VSPHERE_PASSWORD`        | -                        |
| `--vmwarevsphere-network`         | `VSPHERE_NETWORK`         | -                        |
| `--vmwarevsphere-datastore`       | `VSPHERE_DATASTORE`       | -                        |
| `--vmwarevsphere-datacenter`      | `VSPHERE_DATACENTER`      | -                        |
| `--vmwarevsphere-pool`            | `VSPHERE_POOL`            | -                        |
| `--vmwarevsphere-compute-ip`      | `VSPHERE_COMPUTE_IP`      | -                        |

## Specify a base operating systems

The Machine provisioning system supports several base operating systems. For
local providers such as VirtualBox, Fusion, Hyper-V, and so forth, the default
base operating system is Boot2Docker. For cloud providers, the base operating
system is the latest Ubuntu LTS the provider supports.

| Operating System           | Version          | Notes                   |
|----------------------------|------------------|-------------------------|
| Boot2Docker                | 1.5+             | default for local       |
| Ubuntu                     | 12.04+           | default for remote      |
| RancherOS                  | 0.3+             |                         |
| Debian                     | 8.0+             | experimental            |
| RedHat Enterprise Linux    | 7.0+             | experimental            |
| CentOS                     | 7+               | experimental            |
| Fedora                     | 21+              | experimental            |

To use a different base operating system on a remote provider, specify the
provider's image flag and one of its available images. For example, to
select a `debian-8-x64` image on DigitalOcean you would supply the following:

    --digitalocean-image=debian-8-x64

If you change the base image for a provider, you may also need to change
the SSH user. For example, the default Red Hat AMI on EC2 expects the
SSH user to be `ec2-user`, so you would have to specify this with
`--amazonec2-ssh-user ec2-user`.

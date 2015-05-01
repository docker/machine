---
page_title: Docker Machine
page_description: Working with Docker Machine
page_keywords: docker, machine, amazonec2, azure, digitalocean, google, openstack, rackspace, softlayer, virtualbox, vmwarevcloudair, vmwarevsphere
---


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

Docker Machine is supported on Windows, OSX, and Linux and is installable as one
standalone binary.  The links to the binaries for the various platforms and
architectures are below:

- [Windows - 32bit](https://github.com/docker/machine/releases/download/v0.2.0/docker-machine_windows-386.exe)
- [Windows - 64bit](https://github.com/docker/machine/releases/download/v0.2.0/docker-machine_windows-amd64.exe)
- [OSX - x86_64](https://github.com/docker/machine/releases/download/v0.2.0/docker-machine_darwin-amd64)
- [OSX - (old macs)](https://github.com/docker/machine/releases/download/v0.2.0/docker-machine_darwin-386)
- [Linux - x86_64](https://github.com/docker/machine/releases/download/v0.2.0/docker-machine_linux-amd64)
- [Linux - i386](https://github.com/docker/machine/releases/download/v0.2.0/docker-machine_linux-386)

### OSX and Linux

To install on OSX or Linux, download the proper binary to somewhere in your
`PATH` (e.g. `/usr/local/bin`) and make it executable.  For instance, to install on
most OSX machines these commands should suffice:

```
$ curl -L https://github.com/docker/machine/releases/download/v0.2.0/docker-machine_darwin-amd64 > /usr/local/bin/docker-machine
$ chmod +x /usr/local/bin/docker-machine
```

For Linux, just substitute "linux" for "darwin" in the binary name above.

Now you should be able to check the version with `docker-machine -v`:

```
$ docker-machine -v
machine version 0.2.0
```

In order to run Docker commands on your machines without having to use SSH, make
sure to install the Docker client as well, e.g.:

```
$ curl -L https://get.docker.com/builds/Darwin/x86_64/docker-latest > /usr/local/bin/docker
```

### Windows

Currently, Docker recommends that you install and use Docker Machine on Windows
with [msysgit](https://msysgit.github.io/).  This will provide you with some
programs that Docker Machine relies on such as `ssh`, as well as a functioning
shell.

When you have installed msysgit, start up the terminal prompt and run the
following commands.  Here it is assumed that you are on a 64-bit Windows
installation.  If you are on a 32-bit installation, please substitute "i386" for
"x86_64" in the URLs mentioned.

First, install the Docker client binary:

```
$ curl -L https://get.docker.com/builds/Windows/x86_64/docker-latest.exe > /bin/docker
```

Next, install the Docker Machine binary:

```
$ curl -L https://github.com/docker/machine/releases/download/v0.2.0/docker-machine_windows-amd64.exe > /bin/docker-machine
```

Now running `docker-machine` should work.

```
$ docker-machine -v
machine version 0.2.0
```

## Getting started with Docker Machine using a local VM

Let's take a look at using `docker-machine` for creating, using, and managing a
Docker host inside of [VirtualBox](https://www.virtualbox.org/).

First, ensure that
[VirtualBox 4.3.26](https://www.virtualbox.org/wiki/Downloads) is correctly
installed on your system.

If you run the `docker-machine ls` command to show all available machines, you will see
that none have been created so far.

```
$ docker-machine ls
NAME   ACTIVE   DRIVER   STATE   URL
```

To create one, we run the `docker-machine create` command, passing the string
`virtualbox` to the `--driver` flag.  The final argument we pass is the name of
the machine - in this case, we will name our machine "dev".

This command will download a lightweight Linux distribution
([boot2docker](https://github.com/boot2docker/boot2docker)) with the Docker
daemon installed, and will create and start a VirtualBox VM with Docker running.


```
$ docker-machine create --driver virtualbox dev
INFO[0001] Downloading boot2docker.iso to /home/<your username>/.docker/machine/cache/boot2docker.iso...
INFO[0011] Creating SSH key...
INFO[0012] Creating VirtualBox VM...
INFO[0019] Starting VirtualBox VM...
INFO[0020] Waiting for VM to start...
INFO[0053] To see how to connect Docker to this machine, run: docker-machine env dev"
```

You can see the machine you have created by running the `docker-machine ls` command
again:

```
$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL                         SWARM
dev             virtualbox   Running   tcp://192.168.99.100:2376
```

Next, as noted in the output of the `docker-machine create` command, we have to tell
Docker to talk to that machine.  You can do this with the `docker-machine env`
command.  For example,

```
$ eval "$(docker-machine env dev)"
$ docker ps
```

> **Note**: If you are using `fish`, or a Windows shell such as
> Powershell/`cmd.exe` the above method will not work as described.  Instead,
> see [the `env` command's documentation](https://docs.docker.com/machine/#env)
> to learn how to set the environment variables for your shell.

This will set environment variables that the Docker client will read which specify
the TLS settings. Note that you will need to do that every time you open a new tab or
restart your machine.

To see what will be set, run `docker-machine env dev`.

```
$ docker-machine env dev
export DOCKER_TLS_VERIFY=1
export DOCKER_CERT_PATH=/Users/<your username>/.docker/machine/machines/dev
export DOCKER_HOST=tcp://192.168.99.100:2376
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
$ docker-machine ip
192.168.99.100
```

For instance, you can try running a webserver ([nginx](https://nginx.org)) in a
container with the following command:

```
$ docker run -d -p 8000:80 nginx
```

When the image is finished pulling, you can hit the server at port 8000 on the
IP address given to you by `docker-machine ip`.  For instance:

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
run `docker-machine create` again.  All created machines will appear in the
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
the only thing Docker Machine is capable of.  Docker Machine supports several
“drivers” which let you use the same interface to create hosts on many different
cloud or local virtualization platforms.  This is accomplished by using the
`docker-machine create` command with the `--driver` flag.  Here we will be
demonstrating the [Digital Ocean](https://digitalocean.com) driver (called
`digitalocean`), but there are drivers included for several providers including
Amazon Web Services, Google Compute Engine, and Microsoft Azure.

Usually it is required that you pass account verification credentials for these
providers as flags to `docker-machine create`.  These flags are unique for each driver.
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
INFO[0000] Creating SSH key...
INFO[0000] Creating Digital Ocean droplet...
INFO[0002] Waiting for SSH...
INFO[0085] "staging" has been created.
INFO[0085] To see how to connect Docker to this machine, run: docker-machine env staging"
```

For convenience, `docker-machine` will use sensible defaults for choosing
settings such as the image that the VPS is based on, but they can also be
overridden using their respective flags (e.g. `--digitalocean-image`).  This is
useful if, for instance, you want to create a nice large instance with a lot of
memory and CPUs (by default `docker-machine` creates a small VPS).  For a full
list of the flags/settings available and their defaults, see the output of
`docker-machine create -h`.

When the creation of a host is initiated, a unique SSH key for accessing the
host (initially for provisioning, then directly later if the user runs the
`docker-machine ssh` command) will be created automatically and stored in the
client's directory in `~/.docker/machines`.  After the creation of the SSH key,
Docker will be installed on the remote machine and the daemon will be configured
to accept remote connections over TCP using TLS for authentication.  Once this
is finished, the host is ready for connection.

To prepare the Docker client to send commands to the remote server we have
created, we can use the subshell method again:

```
$ eval "$(docker-machine env staging)"
```

From this point, the remote host behaves much like the local host we created in
the last section.  If we look at `docker-machine ls`, we'll see it is now the
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

> **Note**: This is an experimental feature so the subcommands and
> options are likely to change in future versions.

First, create a Swarm token.  Optionally, you can use another discovery service.
See the Swarm docs for details.

To create the token, first create a Machine.  This example will use VirtualBox.

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

### Swarm Master

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

### Swarm Nodes

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
INFO[0001] Downloading boot2docker.iso to /home/ehazlett/.docker/machine/cache/boot2docker.iso...
INFO[0000] Creating SSH key...
INFO[0000] Creating VirtualBox VM...
INFO[0007] Starting VirtualBox VM...
INFO[0007] Waiting for VM to start...
INFO[0038] To see how to connect Docker to this machine, run: docker-machine env dev
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
are working with.  To that extent, specifying an argument to the `-d` flag will
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
   --engine-flag [--engine-flag option --engine-flag option]                                            Specify arbitrary flags to include with the created engine in the form flag=value
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
configures it with some sensible defaults.  For instance, it allows connection
from the outside world over TCP with TLS-based encryption and defaults to AUFS
as the [storage
driver](https://docs.docker.com/reference/commandline/cli/#daemon-storage-driver-option)
when available.

There are several cases where the user might want to set options for the created
Docker engine (also known as the Docker _daemon_) themselves.  For example, they
may want to allow connection to a [registry](https://docs.docker.com/registry/)
that they are running themselves using the `--insecure-registry` flag for the
daemon.  Docker Machine supports the configuration of such options for the
created engines via the `create` command flags which begin with `--engine`.

Note that Docker Machine simply sets the configured parameters on the daemon
and does not set up any of the "dependencies" for you.  For instance, if you
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
insecure registry located at `registry.myco.com`.  You can verify much of this
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
Machine also supports an additional flag, `--engine-flag`, which can be used to
specify arbitrary daemon options with the syntax `--engine-flag
flagname=value`.  For example, to specify that the daemon should use `8.8.8.8`
as the DNS server for all containers, and always use the `syslog` [log
driver](https://docs.docker.com/reference/run/#logging-drivers-log-driver) you
could run the following create command:

```
$ docker-machine create -d virtualbox \
    --engine-flag dns=8.8.8.8 \
    --engine-flag log-driver=syslog \
    gdns
```

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
run in a subshell.  Running `docker-machine env -u` will print `unset` commands
which reverse this effect.

```
$ env | grep DOCKER
$ eval "$(docker-machine env dev)"
$ env | grep DOCKER
DOCKER_HOST=tcp://192.168.99.101:2376
DOCKER_CERT_PATH=/Users/nathanleclaire/.docker/machines/.client
DOCKER_TLS_VERIFY=1
$ # If you run a docker command, now it will run against that host.
$ eval "$(docker-machine env -u)"
$ env | grep DOCKER
$ # The environment variables have been unset.
```

The output described above is intended for the shells `bash` and `zsh` (if
you're not sure which shell you're using, there's a very good possibility that
it's `bash`).  However, these are not the only shells which Docker Machine
supports.

If you are using `fish` and the `SHELL` environment variable is correctly set to
the path where `fish` is located, `docker-machine env name` will print out the
values in the format which `fish` expects:

```
set -x DOCKER_TLS_VERIFY 1;
set -x DOCKER_CERT_PATH "/Users/nathanleclaire/.docker/machine/machines/overlay";
set -x DOCKER_HOST tcp://192.168.99.102:2376;
# Run this command to configure your shell: eval (docker-machine env overlay)
```

If you are on Windows and using Powershell or `cmd.exe`, `docker-machine env`
cannot detect your shell automatically, but it does have support for these
shells.  In order to use them, specify which shell you would like to print the
options for using the `--shell` flag for `docker-machine env`.

For Powershell:

```
$ docker-machine.exe env --shell powershell dev
$Env:DOCKER_TLS_VERIFY = "1"
$Env:DOCKER_HOST = "tcp://192.168.99.101:2376"
$Env:DOCKER_CERT_PATH = "C:\Users\captain\.docker\machine\machines\dev"
# Run this command to configure your shell: docker-machine.exe env --shell=powershell | Invoke-Expression
```

For `cmd.exe`:

```
$ docker-machine.exe env --shell cmd dev
set DOCKER_TLS_VERIFY=1
set DOCKER_HOST=tcp://192.168.99.101:2376
set DOCKER_CERT_PATH=C:\Users\captain\.docker\machine\machines\dev
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
$ docker-machine ip
192.168.99.104
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

List machines.

```
$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL
dev             virtualbox   Stopped
foo0            virtualbox   Running   tcp://192.168.99.105:2376
foo1            virtualbox   Running   tcp://192.168.99.106:2376
foo2            virtualbox   Running   tcp://192.168.99.107:2376
foo3            virtualbox   Running   tcp://192.168.99.108:2376
foo4   *        virtualbox   Running   tcp://192.168.99.109:2376
```

#### regenerate-certs

Regenerate TLS certificates and update the machine with new certs.

```
$ docker-machine regenerate-certs
Regenerate TLS machine certs?  Warning: this is irreversible. (y/n): y
INFO[0013] Regenerating TLS certificates
```

#### restart

Restart a machine.  Oftentimes this is equivalent to
`docker-machine stop; machine start`.

```
$ docker-machine restart
INFO[0005] Waiting for VM to start...
```

#### rm

Remove a machine.  This will remove the local reference as well as delete it
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

#### start

Gracefully start a machine.

```
$ docker-machine restart
INFO[0005] Waiting for VM to start...
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

Upgrade a machine to the latest version of Docker.  If the machine uses Ubuntu
as the underlying operating system, it will upgrade the package `lxc-docker`
(our recommended install method).  If the machine uses boot2docker, this command
will download the latest boot2docker ISO and replace the machine's existing ISO
with the latest.

```
$ docker-machine upgrade dev
INFO[0000] Stopping machine to do the upgrade...
INFO[0005] Upgrading machine dev...
INFO[0006] Downloading latest boot2docker release to /tmp/store/cache/boot2docker.iso...
INFO[0008] Starting machine back up...
INFO[0008] Waiting for VM to start...
```

> **Note**: If you are using a custom boot2docker ISO specified using
> `--virtualbox-boot2docker-url` or an equivalent flag, running an upgrade on
> that machine will completely replace the specified ISO with the latest
> "vanilla" boot2docker ISO available.

#### url

Get the URL of a host

```
$ docker-machine url
tcp://192.168.99.109:2376
```

## Drivers

#### Amazon Web Services
Create machines on [Amazon Web Services](http://aws.amazon.com).  You will need an Access Key ID, Secret Access Key and a VPC ID.  To find the VPC ID, login to the AWS console and go to Services -> VPC -> Your VPCs.  Select the one where you would like to launch the instance.

Options:

 - `--amazonec2-access-key`: **required** Your access key id for the Amazon Web Services API.
 - `--amazonec2-ami`: The AMI ID of the instance to use  Default: `ami-cc3b3ea4`
 - `--amazonec2-instance-type`: The instance type to run.  Default: `t2.micro`
 - `--amazonec2-iam-instance-profile`: The AWS IAM role name to be used as the instance profile
 - `--amazonec2-region`: The region to use when launching the instance.  Default: `us-east-1`
 - `--amazonec2-root-size`: The root disk size of the instance (in GB).  Default: `16`
 - `--amazonec2-secret-key`: **required** Your secret access key for the Amazon Web Services API.
 - `--amazonec2-security-group`: AWS VPC security group name. Default: `docker-machine`
 - `--amazonec2-session-token`: Your session token for the Amazon Web Services API.
 - `--amazonec2-subnet-id`: AWS VPC subnet id
 - `--amazonec2-vpc-id`: **required** Your VPC ID to launch the instance in.
 - `--amazonec2-zone`: The AWS zone launch the instance in (i.e. one of a,b,c,d,e). Default: `a`
 - `--amazonec2-private-address-only`: Use the private IP address only

By default, the Amazon EC2 driver will use a daily image of Ubuntu 14.04 LTS.

| Region        | AMI ID     |
|:--------------|:-----------|
|ap-northeast-1 |ami-fc11d4fc|
|ap-southeast-1 |ami-7854692a|
|ap-southeast-2 |ami-c5611cff|
|cn-north-1     |ami-7cd84545|
|eu-west-1      |ami-2d96f65a|
|eu-central-1   |ami-3cdae621|
|sa-east-1      |ami-71b2376c|
|us-east-1      |ami-cc3b3ea4|
|us-west-1      |ami-017f9d45|
|us-west-2      |ami-55526765|
|us-gov-west-1  |ami-8ffa9bac|

#### Digital Ocean

Create Docker machines on [Digital Ocean](https://www.digitalocean.com/).

You need to create a personal access token under "Apps & API" in the Digital Ocean
Control Panel and pass that to `docker-machine create` with the `--digitalocean-access-token` option.

    $ docker-machine create --driver digitalocean --digitalocean-access-token=aa9399a2175a93b17b1c86c807e08d3fc4b79876545432a629602f61cf6ccd6b test-this

Options:

 - `--digitalocean-access-token`: Your personal access token for the Digital Ocean API.
 - `--digitalocean-image`: The name of the Digital Ocean image to use. Default: `docker`
 - `--digitalocean-region`: The region to create the droplet in, see [Regions API](https://developers.digitalocean.com/documentation/v2/#regions) for how to get a list. Default: `nyc3`
 - `--digitalocean-size`: The size of the Digital Ocean droplet (larger than default options are of the form `2gb`). Default: `512mb`
 - `--digitalocean-ipv6`: Enable IPv6 support for the droplet. Default: `false`
 - `--digitalocean-private-networking`: Enable private networking support for the droplet. Default: `false`
 - `--digitalocean-backups`: Enable Digital Oceans backups for the droplet. Default: `false`

The DigitalOcean driver will use `ubuntu-14-04-x64` as the default image.

#### exoscale
Create machines on [exoscale](https://www.exoscale.ch/).

Get your API key and API secret key from [API details](https://portal.exoscale.ch/account/api) and pass them to `machine create` with the `--exoscale-api-key` and `--exoscale-api-secret-key` options.

Options:

 - `--exoscale-api-key`: Your API key.
 - `--exoscale-api-secret-key`: Your API secret key.
 - `--exoscale-instance-profile`: Instance profile. Default: `small`.
 - `--exoscale-disk-size`: Disk size for the host in GB. Default: `50`.
 - `--exoscale-security-group`: Security group. It will be created if it doesn't exist. Default: `docker-machine`.

If a custom security group is provided, you need to ensure that you allow TCP port 2376 in an ingress rule.

#### Generic
Create machines using an existing VM/Host with SSH.

This is useful if you are using a provider that Machine does not support
directly or if you would like to import an existing host to allow Docker
Machine to manage.

Options:

 - `--generic-ip-address`: IP Address of host
 - `--generic-ssh-user`: SSH username used to connect (default: `root`)
 - `--generic-ssh-key`: Path to the SSH user private key
 - `--generic-ssh-port`: Port to use for SSH (default: `22`)

> Note: you must use a base Operating System supported by Machine.

#### Google Compute Engine
Create machines on [Google Compute Engine](https://cloud.google.com/compute/).  You will need a Google account and project name.  See https://cloud.google.com/compute/docs/projects for details on projects.

The Google driver uses oAuth.  When creating the machine, you will have your browser opened to authorize.  Once authorized, paste the code given in the prompt to launch the instance.

Options:

 - `--google-zone`: The zone to launch the instance.  Default: `us-central1-a`
 - `--google-machine-type`: The type of instance.  Default: `f1-micro`
 - `--google-username`: The username to use for the instance.  Default: `docker-user`
 - `--google-project`: The name of your project to use when launching the instance.
 - `--google-auth-token`: Your oAuth token for the Google Cloud API.
 - `--google-scopes`: The scopes for OAuth 2.0 to Access Google APIs. See [Google Compute Engine Doc](https://cloud.google.com/storage/docs/authentication).
 - `--google-disk-size`: The disk size of instance. Default: `10`
 - `--google-disk-type`: The disk type of instance. Default: `pd-standard`
 
The GCE driver will use the `ubuntu-1404-trusty-v20150316` instance type unless otherwise specified.

#### IBM Softlayer

Create machines on [Softlayer](http://softlayer.com).

You need to generate an API key in the softlayer control panel.
[Retrieve your API key](http://knowledgelayer.softlayer.com/procedure/retrieve-your-api-key)

Options:

  - `--softlayer-api-endpoint`: Change softlayer API endpoint
  - `--softlayer-user`: **required** username for your softlayer account, api key needs to match this user.
  - `--softlayer-api-key`: **required** API key for your user account
  - `--softlayer-cpu`: Number of CPU's for the machine.
  - `--softlayer-disk-size: Size of the disk in MB. `0` sets the softlayer default.
  - `--softlayer-domain`: **required** domain name for the machine
  - `--softlayer-hostname`: hostname for the machine
  - `--softlayer-hourly-billing`: Sets the hourly billing flag (default), otherwise uses monthly billing
  - `--softlayer-image`: OS Image to use
  - `--softlayer-local-disk`: Use local machine disk instead of softlayer SAN.
  - `--softlayer-memory`: Memory for host in MB
  - `--softlayer-private-net-only`: Disable public networking
  - `--softlayer-region`: softlayer region

The SoftLayer driver will use `UBUNTU_LATEST` as the image type by default.


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

Options:

 - `--azure-subscription-id`: Your Azure subscription ID (A GUID like `d255d8d7-5af0-4f5c-8a3e-1545044b861e`).
 - `--azure-subscription-cert`: Your Azure subscription cert.

The Azure driver uses the `b39f27a8b8c64d52b05eac6a62ebad85__Ubuntu-14_04_1-LTS-amd64-server-20140927-en-us-30GB`
image by default. Note, this image is not available in the Chinese regions. In China you should
 specify `b549f4301d0b4295b8e76ceb65df47d4__Ubuntu-14_04_1-LTS-amd64-server-20140927-en-us-30GB`.

You may need to `machine ssh` in to the virtual machine and reboot to ensure that the OS is updated.

#### Microsoft Hyper-V

Creates a Boot2Docker virtual machine locally on your Windows machine
using Hyper-V.  [See here](http://windows.microsoft.com/en-us/windows-8/hyper-v-run-virtual-machines)
for instructions to enable Hyper-V. You will need to use an
Administrator level account to create and manage Hyper-V machines.

> **Note**: You will need an existing virtual switch to use the
> driver.  Hyper-V can share an external network interface (aka
> bridging), see [this blog](http://blogs.technet.com/b/canitpro/archive/2014/03/11/step-by-step-enabling-hyper-v-for-use-on-windows-8-1.aspx).
> If you would like to use NAT, create an internal network, and use
> [Internet Connection
> Sharing](http://www.packet6.com/allowing-windows-8-1-hyper-v-vm-to-work-with-wifi/).

Options:

 - `--hyper-v-boot2docker-location`: Location of a local boot2docker iso to use. Overrides the URL option below.
 - `--hyper-v-boot2docker-url`: The URL of the boot2docker iso. Defaults to the latest available version.
 - `--hyper-v-disk-size`: Size of disk for the host in MB. Defaults to `20000`.
 - `--hyper-v-memory`: Size of memory for the host in MB. Defaults to `1024`. The machine is setup to use dynamic memory.
 - `--hyper-v-virtual-switch`: Name of the virtual switch to use. Defaults to first found.

#### Openstack
Create machines on [Openstack](http://www.openstack.org/software/)

Mandatory:

 - `--openstack-flavor-id` or `openstack-flavor-name`: Identify the flavor that will be used for the machine.
 - `--openstack-image-id` or `openstack-image-name`: Identify the image that will be used for the machine.

Options:

 - `--openstack-auth-url`: Keystone service base URL.
 - `--openstack-domain-name` or `--openstack-domain-id`: Domain to use for authentication (Keystone v3 only)
 - `--openstack-username`: User identifer to authenticate with.
 - `--openstack-password`: User password. It can be omitted if the standard environment variable `OS_PASSWORD` is set.
 - `--openstack-tenant-name` or `--openstack-tenant-id`: Identify the tenant in which the machine will be created.
 - `--openstack-region`: The region to work on. Can be omitted if there is ony one region on the OpenStack.
 - `--openstack-endpoint-type`: Endpoint type can be `internalURL`, `adminURL` on `publicURL`. If is a helper for the driver
   to choose the right URL in the OpenStack service catalog. If not provided the default id `publicURL`
 - `--openstack-net-id` or `--openstack-net-name`: Identify the private network the machine will be connected on. If your OpenStack project project contains only one private network it will be use automatically.
 - `--openstack-sec-groups`: If security groups are available on your OpenStack you can specify a comma separated list
   to use for the machine (e.g. `secgrp001,secgrp002`).
 - `--openstack-floatingip-pool`: The IP pool that will be used to get a public IP an assign it to the machine. If there is an
   IP address already allocated but not assigned to any machine, this IP will be chosen and assigned to the machine. If
   there is no IP address already allocated a new IP will be allocated and assigned to the machine.
 - `--openstack-ssh-user`: The username to use for SSH into the machine. If not provided `root` will be used.
 - `--openstack-ssh-port`: Customize the SSH port if the SSH server on the machine does not listen on the default port.
 - `--openstack-insecure`: Explicitly allow openstack driver to perform "insecure" SSL (https) requests. The server's certificate will not be verified against any certificate authorities. This option should be used with caution.

Environment variables:

Here comes the list of the supported variables with the corresponding options. If both environment variable
and CLI option are provided the CLI option takes the precedence.

| Environment variable | CLI option                  |
|----------------------|-----------------------------|
| `OS_AUTH_URL`        | `--openstack-auth-url`      |
| `OS_DOMAIN_ID`       | `--openstack-domain-id`     |
| `OS_DOMAIN_NAME`     | `--openstack-domain-name`   |
| `OS_USERNAME`        | `--openstack-username`      |
| `OS_PASSWORD`        | `--openstack-password`      |
| `OS_TENANT_NAME`     | `--openstack-tenant-name`   |
| `OS_TENANT_ID`       | `--openstack-tenant-id`     |
| `OS_REGION_NAME`     | `--openstack-region`        |
| `OS_ENDPOINT_TYPE`   | `--openstack-endpoint-type` |

#### Rackspace
Create machines on [Rackspace cloud](http://www.rackspace.com/cloud)

Options:

 - `--rackspace-username`: Rackspace account username
 - `--rackspace-api-key`: Rackspace API key
 - `--rackspace-region`: Rackspace region name
 - `--rackspace-endpoint-type`: Rackspace endpoint type (adminURL, internalURL or the default publicURL)
 - `--rackspace-image-id`: Rackspace image ID. Default: Ubuntu 14.10 (Utopic Unicorn) (PVHVM)
 - `--rackspace-flavor-id`: Rackspace flavor ID. Default: General Purpose 1GB
 - `--rackspace-ssh-user`: SSH user for the newly booted machine. Set to root by default
 - `--rackspace-ssh-port`: SSH port for the newly booted machine. Set to 22 by default

Environment variables:

Here comes the list of the supported variables with the corresponding options. If both environment
variable and CLI option are provided the CLI option takes the precedence.

| Environment variable | CLI option                  |
|----------------------|-----------------------------|
| `OS_USERNAME`        | `--rackspace-username`      |
| `OS_API_KEY`         | `--rackspace-api-key`       |
| `OS_REGION_NAME`     | `--rackspace-region`        |
| `OS_ENDPOINT_TYPE`   | `--rackspace-endpoint-type` |
| `OS_FLAVOR_ID`       | `--rackspace-flavor-id`     |

The Rackspace driver will use `598a4282-f14b-4e50-af4c-b3e52749d9f9` (Ubuntu 14.04 LTS) by default.

#### Oracle VirtualBox

Create machines locally using [VirtualBox](https://www.virtualbox.org/).
This driver requires VirtualBox to be installed on your host.

    $ docker-machine create --driver=virtualbox vbox-test

Options:

 - `--virtualbox-boot2docker-url`: The URL of the boot2docker image. Defaults to the latest available version.
 - `--virtualbox-disk-size`: Size of disk for the host in MB. Default: `20000`
 - `--virtualbox-memory`: Size of memory for the host in MB. Default: `1024`
 - `--virtualbox-cpu-count`: Number of CPUs to use to create the VM. Defaults to single CPU.

The `--virtualbox-boot2docker-url` flag takes a few different forms.  By
default, if no value is specified for this flag, Machine will check locally for
a boot2docker ISO.  If one is found, that will be used as the ISO for the
created machine.  If one is not found, the latest ISO release available on
[boot2docker/boot2docker](https://github.com/boot2docker/boot2docker) will be
downloaded and stored locally for future use.  Note that this means you must run
`docker-machine upgrade` deliberately on a machine if you wish to update the "cached"
boot2docker ISO.

This is the default behavior (when `--virtualbox-boot2docker-url=""`), but the
option also supports specifying ISOs by the `http://` and `file://` protocols.
`file://` will look at the path specified locally to locate the ISO: for
instance, you could specify `--virtualbox-boot2docker-url
file://$HOME/Downloads/rc.iso` to test out a release candidate ISO that you have
downloaded already.  You could also just get an ISO straight from the Internet
using the `http://` form.

Environment variables:

Here comes the list of the supported variables with the corresponding options. If both environment
variable and CLI option are provided the CLI option takes the precedence.

| Environment variable              | CLI option                        |
|-----------------------------------|-----------------------------------|
| `VIRTUALBOX_MEMORY_SIZE`          | `--virtualbox-memory`             |
| `VIRTUALBOX_CPU_COUNT`            | `--virtualbox-cpu-count`          |
| `VIRTUALBOX_DISK_SIZE`            | `--virtualbox-disk-size`          |
| `VIRTUALBOX_BOOT2DOCKER_URL`      | `--virtualbox-boot2docker-url`    |


#### VMware Fusion
Creates machines locally on [VMware Fusion](http://www.vmware.com/products/fusion). Requires VMware Fusion to be installed.

Options:

 - `--vmwarefusion-boot2docker-url`: URL for boot2docker image.
 - `--vmwarefusion-disk-size`: Size of disk for host VM (in MB). Default: `20000`
 - `--vmwarefusion-memory-size`: Size of memory for host VM (in MB). Default: `1024`

The VMware Fusion driver uses the latest boot2docker image.

#### VMware vCloud Air
Creates machines on [vCloud Air](http://vcloud.vmware.com) subscription service. You need an account within an existing subscription of vCloud Air VPC or Dedicated Cloud.

Options:

 - `--vmwarevcloudair-username`: vCloud Air Username.
 - `--vmwarevcloudair-password`: vCloud Air Password.
 - `--vmwarevcloudair-catalog`: Catalog. Default: `Public Catalog`
 - `--vmwarevcloudair-catalogitem`: Catalog Item. Default: `Ubuntu Server 12.04 LTS (amd64 20140927)`
 - `--vmwarevcloudair-computeid`: Compute ID (if using Dedicated Cloud).
 - `--vmwarevcloudair-cpu-count`: VM Cpu Count. Default: `1`
 - `--vmwarevcloudair-docker-port`: Docker port. Default: `2376`
 - `--vmwarevcloudair-edgegateway`: Organization Edge Gateway. Default: `<vdcid>`
 - `--vmwarevcloudair-memory-size`: VM Memory Size in MB. Default: `2048`
 - `--vmwarevcloudair-name`: vApp Name. Default: `<autogenerated>`
 - `--vmwarevcloudair-orgvdcnetwork`: Organization VDC Network to attach. Default: `<vdcid>-default-routed`
 - `--vmwarevcloudair-provision`: Install Docker binaries. Default: `true`
 - `--vmwarevcloudair-publicip`: Org Public IP to use.
 - `--vmwarevcloudair-ssh-port`: SSH port. Default: `22`
 - `--vmwarevcloudair-vdcid`: Virtual Data Center ID.

The VMware vCloud Air driver will use the `Ubuntu Server 12.04 LTS (amd64 20140927)` image by default.

#### VMware vSphere
Creates machines on a [VMware vSphere](http://www.vmware.com/products/vsphere) Virtual Infrastructure. Requires a working vSphere (ESXi and optionally vCenter) installation. The vSphere driver depends on [`govc`](https://github.com/vmware/govmomi/tree/master/govc) (must be in path) and has been tested with [vmware/govmomi@`c848630`](https://github.com/vmware/govmomi/commit/c8486300bfe19427e4f3226e3b3eac067717ef17).

Options:

 - `--vmwarevsphere-username`: vSphere Username.
 - `--vmwarevsphere-password`: vSphere Password.
 - `--vmwarevsphere-boot2docker-url`: URL for boot2docker image.
 - `--vmwarevsphere-compute-ip`: Compute host IP where the Docker VM will be instantiated.
 - `--vmwarevsphere-cpu-count`: CPU number for Docker VM. Default: `2`
 - `--vmwarevsphere-datacenter`: Datacenter for Docker VM (must be set to `ha-datacenter` when connecting to a single host).
 - `--vmwarevsphere-datastore`: Datastore for Docker VM.
 - `--vmwarevsphere-disk-size`: Size of disk for Docker VM (in MB). Default: `20000`
 - `--vmwarevsphere-memory-size`: Size of memory for Docker VM (in MB). Default: `2048`
 - `--vmwarevsphere-network`: Network where the Docker VM will be attached.
 - `--vmwarevsphere-pool`: Resource pool for Docker VM.
 - `--vmwarevsphere-vcenter`: IP/hostname for vCenter (or ESXi if connecting directly to a single host).

The VMware vSphere driver uses the latest boot2docker image.

## Base Operating Systems
The default base operating system for Machine is Boot2Docker on local providers
(VirtualBox, Fusion, Hyper-V, etc) and the latest Ubuntu LTS supported
by the cloud provider.  RedHat Enterprise Linux is also supported.  To use
RHEL, you will need to select the image accordingly with the provider.  For
example, in Amazon EC2, you could use "ami-12663b7a" as the
`--amazonec2-ami` option which create an instance using RHEL 7.1 64-bit.

## Release Notes

### Version 0.2.0 (April 16, 2015)

For complete information on this release, see the [0.2.0 Milestone project page](https://github.com/docker/machine/wiki/0.2.0-Milestone-Project-Page).
In addition to bug fixes and refinements, this release adds the following:

* Updated and refactored Driver interface For details, see
[PR #694](https://github.com/docker/machine/pull/694).

* Initial creation of an internal API, so Machine can be used as a library. For
details, see [PR #553](https://github.com/docker/machine/issues/553).

* Improvements and isolation of provisioning functionality, so Machine can
provision and configure the Docker Engine based on OS detection. For details,
see [PR #553](https://github.com/docker/machine/issues/553).

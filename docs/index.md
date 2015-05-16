---
page_title: Docker Machine
page_description: Working with Docker Machine
page_keywords: docker, machine, amazonec2, azure, digitalocean, google, openstack, rackspace, softlayer, virtualbox, vmwarevcloudair, vmwarevsphere
---


# Docker Machine

> **Note**: Machine is currently in beta, so things are likely to change. We
> don't recommend you use it in production yet.

Machine makes it really easy to create Docker hosts on your computer, on cloud
providers and inside your own data center. It creates servers, installs Docker
on them, then configures the Docker client to talk to them.

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

Docker Machine is supported on Windows, OSX, and Linux.  To install Docker
Machine, download the appropriate binary for your OS and architecture, rename it `docker-machine` and place
into your `PATH`:

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

Let's take a look at using `docker-machine` to creating, using, and managing a Docker
host inside of [VirtualBox](https://www.virtualbox.org/).

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

This will download a lightweight Linux distribution
([boot2docker](https://github.com/boot2docker/boot2docker)) with the Docker
daemon installed, and will create and start a VirtualBox VM with Docker running.


```
$ docker-machine create --driver virtualbox dev
INFO[0001] Downloading boot2docker.iso to /home/<your username>/.docker/machine/cache/boot2docker.iso...
INFO[0011] Creating SSH key...
INFO[0012] Creating VirtualBox VM...
INFO[0019] Starting VirtualBox VM...
INFO[0020] Waiting for VM to start...
INFO[0053] "dev" has been created and is now the active machine.
INFO[0053] To point your Docker client at it, run this in your shell: eval "$(docker-machine env dev)"
```

You can see the machine you have created by running the `docker-machine ls` command
again:

```
$ docker-machine ls
NAME   ACTIVE   DRIVER       STATE     URL                         SWARM
dev    *        virtualbox   Running   tcp://192.168.99.100:2376
```

The `*` next to `dev` indicates that it is the active host.

Next, as noted in the output of the `docker-machine create` command, we have to tell
Docker to talk to that machine.  You can do this with the `docker-machine env`
command.  For example,

```
$ eval "$(docker-machine env dev)"
$ docker ps
```

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

Now you can manage as many local VMs running Docker as you please- just run
`docker-machine create` again.

If you are finished using a host, you can stop it with `docker stop` and start
it again with `docker start`:

```
$ docker-machine stop
$ docker-machine start
```

If they aren't passed any arguments, commands such as `docker-machine stop` will run
against the active host (in this case, the VirtualBox VM).  You can also specify
a host to run a command against as an argument.  For instance, you could also
have written:

```
$ docker-machine stop dev
$ docker-machine start dev
```

## Using Docker Machine with a cloud provider

One of the nice things about `docker-machine` is that it provides several “drivers”
which let you use the same interface to create hosts on many different cloud
platforms.  This is accomplished by using the `docker-machine create` command with the
 `--driver` flag.  Here we will be demonstrating the
[Digital Ocean](https://digitalocean.com) driver (called `digitalocean`), but
there are drivers included for several providers including Amazon Web Services,
Google Compute Engine, and Microsoft Azure.

Usually it is required that you pass account verification credentials for these
providers as flags to `docker-machine create`.  These flags are unique for each driver.
For instance, to pass a Digital Ocean access token you use the
`--digitalocean-access-token` flag.

Let's take a look at how to do this.

To generate your access token:

1. Go to the Digital Ocean administrator panel and click on "Apps and API" in
the side panel.
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
INFO[0085] "staging" has been created and is now the active machine
INFO[0085] To point your Docker client at it, run this in your shell: eval "$(docker-machine env staging)"
```

For convenience, `docker-machine` will use sensible defaults for choosing settings such
 as the image that the VPS is based on, but they can also be overridden using
their respective flags (e.g. `--digitalocean-image`).  This is useful if, for
instance, you want to create a nice large instance with a lot of memory and CPUs
(by default `docker-machine` creates a small VPS).  For a full list of the
flags/settings available and their defaults, see the output of
`docker-machine create -h`.

When the creation of a host is initiated, a unique SSH key for accessing the
host (initially for provisioning, then directly later if the user runs the
`docker-machine ssh` command) will be created automatically and stored in the client's
directory in `~/.docker/machines`.  After the creation of the SSH key, Docker
will be installed on the remote machine and the daemon will be configured to
accept remote connections over TCP using TLS for authentication.  Once this
is finished, the host is ready for connection.

And then from this point, the remote host behaves much like the local host we
created in the last section. If we look at `docker-machine`, we’ll see it is now the
active host:

```
$ docker-machine active dev
$ docker-machine ls
NAME      ACTIVE   DRIVER         STATE     URL
dev                virtualbox     Running   tcp://192.168.99.103:2376
staging   *        digitalocean   Running   tcp://104.236.50.118:2376
```

To select an active host, you can use the `docker-machine active` command.

```
$ docker-machine active dev
$ docker-machine ls
NAME      ACTIVE   DRIVER         STATE     URL
dev       *        virtualbox     Running   tcp://192.168.99.103:2376
staging            digitalocean   Running   tcp://104.236.50.118:2376
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

Get or set the active machine.

```
$ docker-machine ls
NAME      ACTIVE   DRIVER         STATE     URL
dev                virtualbox     Running   tcp://192.168.99.103:2376
staging   *        digitalocean   Running   tcp://104.236.50.118:2376
$ docker-machine active dev
$ docker-machine ls
NAME      ACTIVE   DRIVER         STATE     URL
dev       *        virtualbox     Running   tcp://192.168.99.103:2376
staging            digitalocean   Running   tcp://104.236.50.118:2376
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
INFO[0038] "dev" has been created and is now the active machine.
INFO[0038] To point your Docker client at it, run this in your shell: eval "$(docker-machine env dev)"
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
run in a subshell.  Running `docker-machine env -u` will print
`unset` commands which reverse this effect.

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

#### inspect

Inspect information about a machine.

```
$ docker-machine inspect dev
{
    "DriverName": "virtualbox",
    "Driver": {
        "MachineName": "docker-host-128be8d287b2028316c0ad5714b90bcfc11f998056f2f790f7c1f43f3d1e6eda",
        "SSHPort": 55834,
        "Memory": 1024,
        "DiskSize": 20000,
        "Boot2DockerURL": ""
    }
}
```

#### help

Show help text.

#### ip

Get the IP address of a machine.

```
$ docker-machine ip
192.168.99.104
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
 - `--amazonec2-ami`: The AMI ID of the instance to use  Default: `ami-4ae27e22`
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

By default, the Amazon EC2 driver will use a daily image of Ubuntu 14.04 LTS.

| Region        | AMI ID     |
|:--------------|:-----------|
|ap-northeast-1 |ami-44f1e245|
|ap-southeast-1 |ami-f95875ab|
|ap-southeast-2 |ami-890b62b3|
|cn-north-1     |ami-fe7ae8c7|
|eu-west-1      |ami-823686f5|
|eu-central-1   |ami-ac1524b1|
|sa-east-1      |ami-c770c1da|
|us-east-1      |ami-4ae27e22|
|us-west-1      |ami-d1180894|
|us-west-2      |ami-898dd9b9|
|us-gov-west-1  |ami-cf5630ec|

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

#### Google Compute Engine
Create machines on [Google Compute Engine](https://cloud.google.com/compute/).  You will need a Google account and project name.  See https://cloud.google.com/compute/docs/projects for details on projects.

The Google driver uses oAuth.  When creating the machine, you will have your browser opened to authorize.  Once authorized, paste the code given in the prompt to launch the instance.

Options:

 - `--google-zone`: The zone to launch the instance.  Default: `us-central1-a`
 - `--google-machine-type`: The type of instance.  Default: `f1-micro`
 - `--google-username`: The username to use for the instance.  Default: `docker-user`
 - `--google-instance-name`: The name of the instance.  Default: `docker-machine`
 - `--google-project`: The name of your project to use when launching the instance.

The GCE driver will use the `ubuntu-1404-trusty-v20150128` instance type unless otherwise specified.

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

At the time of writing the below is a complete list of options, use docker-machine create --help to retrieve the currently available options.

 - `--azure-docker-port "2376"`: Azure Docker port
 - `--azure-image`: Azure image name. Default is Ubuntu 14.04 LTS x64 [$AZURE_IMAGE]
 - `--azure-location "West US"`: Azure location [$AZURE_LOCATION]
 - `--azure-password`: Azure user password
 - `--azure-publish-settings-file`: Azure publish settings file [$AZURE_PUBLISH_SETTINGS_FILE]
 - `--azure-size "Small"`: Azure size [$AZURE_SIZE]
 - `--azure-ssh-port "22"`: Azure SSH port
 - `--azure-subscription-cert`: **required** Azure subscription cert [$AZURE_SUBSCRIPTION_CERT]
 - `--azure-subscription-id`: **required** Azure subscription ID [$AZURE_SUBSCRIPTION_ID]
 - `--azure-username "ubuntu"`: Azure username

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

 - `--openstack-flavor-id`: The flavor ID to use when creating the machine
 - `--openstack-image-id`: The image ID to use when creating the machine.

Options:

 - `--openstack-auth-url`: Keystone service base URL.
 - `--openstack-username`: User identifer to authenticate with.
 - `--openstack-password`: User password. It can be omitted if the standard environment variable `OS_PASSWORD` is set.
 - `--openstack-tenant-name` or `--openstack-tenant-id`: Identify the tenant in which the machine will be created.
 - `--openstack-region`: The region to work on. Can be omitted if there is ony one region on the OpenStack.
 - `--openstack-endpoint-type`: Endpoint type can be `internalURL`, `adminURL` on `publicURL`. If is a helper for the driver
   to choose the right URL in the OpenStack service catalog. If not provided the default id `publicURL`
 - `--openstack-net-id`: The private network id the machine will be connected on. If your OpenStack project project
   contains only one private network it will be use automatically.
 - `--openstack-sec-groups`: If security groups are available on your OpenStack you can specify a comma separated list
   to use for the machine (e.g. `secgrp001,secgrp002`).
 - `--openstack-floatingip-pool`: The IP pool that will be used to get a public IP an assign it to the machine. If there is an
   IP address already allocated but not assigned to any machine, this IP will be chosen and assigned to the machine. If
   there is no IP address already allocated a new IP will be allocated and assigned to the machine.
 - `--openstack-ssh-user`: The username to use for SSH into the machine. If not provided `root` will be used.
 - `--openstack-ssh-port`: Customize the SSH port if the SSH server on the machine does not listen on the default port.

Environment variables:

Here comes the list of the supported variables with the corresponding options. If both environment variable
and CLI option are provided the CLI option takes the precedence.

| Environment variable | CLI option                  |
|----------------------|-----------------------------|
| `OS_AUTH_URL`        | `--openstack-auth-url`      |
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

The Rackspace driver will use `598a4282-f14b-4e50-af4c-b3e52749d9f9` (Ubuntu 14.04 LTS) by default.

#### Oracle VirtualBox

Create machines locally using [VirtualBox](https://www.virtualbox.org/).
This driver requires VirtualBox to be installed on your host.

    $ docker-machine create --driver=virtualbox vbox-test

Options:

 - `--virtualbox-boot2docker-url`: The URL of the boot2docker image. Defaults to the latest available version.
 - `--virtualbox-disk-size`: Size of disk for the host in MB. Default: `20000`
 - `--virtualbox-memory`: Size of memory for the host in MB. Default: `1024`
 - `--virtualbox-cpu-count`: Number of CPUs to use to create the VM. Defaults to number of available CPUs.

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

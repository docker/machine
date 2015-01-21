page_title: Working with Docker Machine
page_description: Working with Docker Machine
page_keywords: docker, machine, virtualbox, digitalocean, amazonec2
 
# Working with Docker Machine

## Overview

In order to run Docker containers, you must have a 
[Docker daemon](https://docs.docker.com/arch) running somewhere. If you’re on a 
Linux system and you want to run a container on your local machine, this is 
straightforward: you run the daemon on your local machine and communicate with 
it over the Unix socket located at `/var/run/docker.sock` (this all happens 
behind the scenes when you run `docker` on the command line). However, if you 
want to control containers from Mac OSX / Windows or manage them on a remote 
server, you’ll need to create a new machine (probably a virtual machine) with 
Docker installed and execute Docker commands for that host remotely.  
Traditionally, the way to do this was either:

- manual (open the web interface or virtualization application, make the machine 
yourself, manually install Docker, etc.) and therefore tedious and error-prone
- with existing automation technologies, which usually entail a quite high skill 
threshold

Docker's [`machine`](https://github.com/docker/machine) is a tool for making the 
process of creating and managing those machines (and running Docker commands 
against them) much faster and easier for users.  `machine` allows users to 
quickly create running instances of the Docker daemon on local virtualization 
platforms (e.g. Virtualbox) or on cloud providers (e.g. AWS EC2) that they can 
connect to and control from their local Docker client binary.

## Installation

Docker Machine is supported on Windows, OSX, and Linux.  To install Docker 
Machine, download the appropriate binary for your OS and architecture to the 
correct place in your `PATH`:

- [Windows - x86_64]()
- [OSX - x86_64]()
- [Linux - x86_64]()
- [Windows - i386]()
- [OSX - i386]()
- [Linux - i386]()

Now you should be able to check the version with `machine -v`:

```
$ machine -v
machine version 0.0.3
```

## Getting started with Docker Machine using a local VM

Let's take a look at using `machine` to creating, using, and managing a Docker 
host inside of [VirtualBox](ihttps://www.virtualbox.org/).

First, ensure that 
[VirtualBox 4.3.20](https://www.virtualbox.org/wiki/Downloads) is correctly 
installed on your system. 

If you run the `machine ls` command to show all available machines, you will see 
that none have been created so far.

```
$ machine ls
NAME   ACTIVE   DRIVER   STATE   URL
```

To create one, we run the `machine create` command, passing the string 
`virtualbox` to the `--driver` flag.  The final argument we pass is the name of 
the machine - in this case, we will name our machine "dev".

This will download a lightweight Linux distribution 
([boot2docker](https://github.com/boot2docker/boot2docker)) with the Docker 
daemon installed, and will create and start a VirtualBox VM with Docker running.


```
$ machine create --driver virtualbox dev
INFO[0000] Creating SSH key...
INFO[0000] Creating VirtualBox VM...
INFO[0007] Starting VirtualBox VM...
INFO[0007] Waiting for VM to start...
INFO[0038] "dev" has been created and is now the active machine
INFO[0038] To connect: docker $(machine config dev) ps
```

You can see the machine you have created by running the `machine ls` command 
again:

```
$ machine ls
NAME      ACTIVE   DRIVER       STATE     URL
dev       *        virtualbox   Running   tcp://192.168.99.100:2375
```

The `*` next to `dev` indicates that it is the active host.

Next, as noted in the output of the `machine create` command, we have to tell 
Docker to talk to that machine.  You can do this with the `machine config`
command.  For example,

```
$ docker $(machine config dev) ps
```

This will pass arguments to the Docker client that specify the TLS settings.
To see what will be passed, run `machine config dev`.

You can now run Docker commands on this host:

```
$ docker $(machine config dev) run busybox echo hello world
Unable to find image 'busybox' locally
Pulling repository busybox
e72ac664f4f0: Download complete
511136ea3c5a: Download complete
df7546f9f060: Download complete
e433a6c5b276: Download complete
hello world
```

Any exposed ports are available on the Docker host’s IP address, which you can 
get using the `machine ip` command:

```
$ machine ip
192.168.99.100
```

Now you can manage as many local VMs running Docker as you please- just run 
`machine create` again.

If you are finished using a host, you can stop it with `docker stop` and start 
it again with `docker start`:

```
$ machine stop
$ machine start
```

If they aren't passed any arguments, commands such as `machine stop` will run 
against the active host (in this case, the VirtualBox VM).  You can also specify 
a host to run a command against as an argument.  For instance, you could also 
have written:

```
$ machine stop dev
$ machine start dev
```

## Using Docker Machine with a cloud provider

One of the nice things about `machine` is that it provides several “drivers” 
which let you use the same interface to create hosts on many different cloud 
platforms.  This is accomplished by using the `machine create` command with the
 `--driver` flag.  Here we will be demonstrating the 
[Digital Ocean](https://digitalocean.com) driver (called `digitalocean`), but 
there are drivers included for several providers including Amazon Web Services, 
Google Compute Engine, and Microsoft Azure.

Usually it is required that you pass account verification credentials for these 
providers as flags to `machine create`.  These flags are unique for each driver.  
For instance, to pass a Digital Ocean access token you use the 
`--digitalocean-access-token` flag.

Let's take a look at how to do this.

To generate your access token:

1. Go to the Digital Ocean administrator panel and click on "Apps and API" in 
the side panel.
2. Click on "Generate New Token".
3. Give the token a clever name (e.g. "machine"), make sure the "Write" checkbox 
is checked, and click on "Generate Token".
4. Grab the big long hex string that is generated (this is your token) and store it somehwere safe.

Now, run `machine create` with the `digitalocean` driver and pass your key to 
the `--digitalocean-access-token` flag.

Example:

```
$ machine create \
    --driver digitalocean \
    --digitalocean-access-token 0ab77166d407f479c6701652cee3a46830fef88b8199722b87821621736ab2d4 \
    staging
INFO[0000] Creating SSH key...
INFO[0000] Creating Digital Ocean droplet...
INFO[0002] Waiting for SSH...
INFO[0085] "staging" has been created and is now the active machine
INFO[0085] To connect: docker $(machine config dev) staging
```

For convenience, `machine` will use sensible defaults for choosing settings such
 as the image that the VPS is based on, but they can also be overridden using 
their respective flags (e.g. `--digitalocean-image`).  This is useful if, for 
instance, you want to create a nice large instance with a lot of memory and CPUs 
(by default `machine` creates a small VPS).  For a full list of the 
flags/settings available and their defaults, see the output of 
`machine create -h`.

When the creation of a host is initiated, a unique SSH key for accessing the 
host (initially for provisioning, then directly later if the user runs the 
`machine ssh` command) will be created automatically and stored in the client's 
directory in `~/.docker/machines`.  After the creation of the SSH key, Docker 
will be installed on the remote machine and the daemon will be configured to 
accept remote connections over TCP using TLS for authentication.  Once this 
is finished, the host is ready for connection.

And then from this point, the remote host behaves much like the local host we 
created in the last section. If we look at `machine`, we’ll see it is now the 
active host:

```
$ machine active dev
$ machine ls
NAME      ACTIVE   DRIVER         STATE     URL
dev                virtualbox     Running   tcp://192.168.99.103:2375
staging   *        digitalocean   Running   tcp://104.236.50.118:2375
```

To select an active host, you can use the `machine active` command.

```
$ machine active dev
$ machine ls
NAME      ACTIVE   DRIVER         STATE     URL
dev       *        virtualbox     Running   tcp://192.168.99.103:2375
staging            digitalocean   Running   tcp://104.236.50.118:2375
```

To remove a host and all of its containers and images, use `machine rm`:

```
$ machine rm dev staging
$ machine ls
NAME      ACTIVE   DRIVER       STATE     URL
```

## Adding a host without a driver

You can add a host to Docker which only has a URL and no driver. Therefore it 
can be used an alias for an existing host so you don’t have to type out the URL 
every time you run a Docker command.

```
$ machine create --url=tcp://50.134.234.20:2376 custombox
$ machine ls
NAME        ACTIVE   DRIVER    STATE     URL
custombox   *        none      Running   tcp://50.134.234.20:2376
```

## Subcommands 

#### active

Get or set the active machine.

```
$ machine ls
NAME      ACTIVE   DRIVER         STATE     URL
dev                virtualbox     Running   tcp://192.168.99.103:2375
staging   *        digitalocean   Running   tcp://104.236.50.118:2375
$ machine active dev
$ machine ls
NAME      ACTIVE   DRIVER         STATE     URL
dev       *        virtualbox     Running   tcp://192.168.99.103:2375
staging            digitalocean   Running   tcp://104.236.50.118:2375
```

#### create

Create a machine.

```
$ machine create --driver virtualbox dev
INFO[0000] Creating SSH key...
INFO[0000] Creating VirtualBox VM...
INFO[0007] Starting VirtualBox VM...
INFO[0007] Waiting for VM to start...
INFO[0038] "dev" has been created and is now the active machine. To point Docker at this machine, run: export DOCKER_HOST=$(machine url) DOCKER_AUTH=identity
```

#### config

Show the Docker client configuration for a machine.

```
$ machine config dev
--tls --tlscacert=/Users/ehazlett/.docker/machines/dev/ca.pem --tlscert=/Users/ehazlett/.docker/machines/dev/cert.pem --tlskey=/Users/ehazlett/.docker/machines/dev/key.pem -H tcp://192.168.99.103:2376
```

#### inspect

Inspect information about a machine.

```
$ machine inspect dev
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
$ machine ip
192.168.99.104
```

#### kill

Kill (abruptly force stop) a machine.

```
$ machine ls
NAME   ACTIVE   DRIVER       STATE     URL
dev    *        virtualbox   Running   tcp://192.168.99.104:2376
$ machine kill dev
$ machine ls
NAME   ACTIVE   DRIVER       STATE     URL
dev    *        virtualbox   Stopped
```

#### ls

List machines.

```
$ machine ls
NAME   ACTIVE   DRIVER       STATE     URL
dev             virtualbox   Stopped
foo0            virtualbox   Running   tcp://192.168.99.105:2376
foo1            virtualbox   Running   tcp://192.168.99.106:2376
foo2            virtualbox   Running   tcp://192.168.99.107:2376
foo3            virtualbox   Running   tcp://192.168.99.108:2376
foo4   *        virtualbox   Running   tcp://192.168.99.109:2376
```

#### restart

Restart a machine.  Oftentimes this is equivalent to 
`machine stop; machine start`.

```
$ machine restart
INFO[0005] Waiting for VM to start...
```

#### rm

Remove a machine.  This will remove the local reference as well as delete it 
on the cloud provider or virtualization management platform.

```
$ machine ls
NAME   ACTIVE   DRIVER       STATE     URL
foo0            virtualbox   Running   tcp://192.168.99.105:2376
foo1            virtualbox   Running   tcp://192.168.99.106:2376
$ machine rm foo1
$ machine ls
NAME   ACTIVE   DRIVER       STATE     URL
foo0            virtualbox   Running   tcp://192.168.99.105:2376
```

#### ssh

Log into or run a command on a machine using SSH.

```
$ machine ssh -c "echo this process ran on a remote machine"
this process ran on a remote machine
$ machine ssh
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

#### start

Gracefully start a machine.

```
$ machine restart
INFO[0005] Waiting for VM to start...
```

#### stop

Gracefully stop a machine.

```
$ machine ls
NAME   ACTIVE   DRIVER       STATE     URL
dev    *        virtualbox   Running   tcp://192.168.99.104:2376
$ machine stop dev
$ machine ls
NAME   ACTIVE   DRIVER       STATE     URL
dev    *        virtualbox   Stopped
```

#### upgrade

Upgrade a machine to the latest version of Docker.

```
$ machine upgrade dev
```

#### url

Get the URL of a host

```
$ machine url
tcp://192.168.99.109:2376
```

## Driver Options

TODO: List all possible values (where applicable) for all flags for every 
driver.

#### Amazon Web Services
Create machines on [Amazon Web Services](http://aws.amazon.com).  You will need an Access Key ID, Secret Access Key and a VPC ID.  To find the VPC ID, login to the AWS console and go to Services -> VPC -> Your VPCs.  Select the one where you would like to launch the instance.

Options:

 - `--amazonec2-access-key`: Your access key id for the Amazon Web Services API.
 - `--amazonec2-ami`: The AMI ID of the instance to use  Default: `ami-a00461c8`
 - `--amazonec2-instance-type`: The instance type to run.  Default: `t2.micro`
 - `--amazonec2-region`: The region to use when launching the instance.  Default: `us-east-1`
 - `--amazonec2-root-size`: The root disk size of the instance (in GB).  Default: `16`
 - `--amazonec2-secret-key`: Your secret access key for the Amazon Web Services API.
 - `--amazonec2-session-token`: Your session token for the Amazon Web Services API.
 - `--amazonec2-vpc-id`: Your VPC ID to launch the instance in.
 - `--amazonec2-zone`: The AWS zone launch the instance in (i.e. one of a,b,c,d,e).

#### Digital Ocean
Creates machines on [Digital Ocean](https://www.digitalocean.com/). You need to create a personal access token under "Apps & API" in the Digital Ocean Control Panel and pass that to `machine create` with the `--digitalocean-access-token` option.

Options:

 - `--digitalocean-access-token`: Your personal access token for the Digital Ocean API.
 - `--digitalocean-image`: The name of the Digital Ocean image to use. Default: `docker`
 - `--digitalocean-region`: The region to create the droplet in. Default: `nyc3`
 - `--digitalocean-size`: The size of the Digital Ocean driver. Default: `512mb`

#### Google Compute Engine
Create machines on [Google Compute Engine](https://cloud.google.com/compute/).  You will need a Google account and project name.  See https://cloud.google.com/compute/docs/projects for details on projects.

The Google driver uses oAuth.  When creating the machine, you will have your browser opened to authorize.  Once authorized, paste the code given in the prompt to launch the instance.

Options:

 - `--google-zone`: The zone to launch the instance.  Default: `us-central1-a`
 - `--google-machine-type`: The type of instance.  Default: `f1-micro`
 - `--google-username`: The username to use for the instance.  Default: `docker-user`
 - `--google-instance-name`: The name of the instance.  Default: `docker-machine`
 - `--google-project`: The name of your project to use when launching the instance.

#### IBM Softlayer

Create machines on [Softlayer](http://softlayer.com).

You need to generate an API key in the softlayer control panel.  
[Retrieve your API key](http://knowledgelayer.softlayer.com/procedure/retrieve-your-api-key)

Options:
  - `--softlayer-api-endpoint=`: Change softlayer API endpoint
  - `--softlayer-user`: **required** username for your softlayer account, api key needs to match this user.
  - `--softlayer-api-key`: **required** API key for your user account
  - `--softlayer-cpu`: Number of CPU's for the machine.
  - `--softlayer-disk-size: Size of the disk in MB. `0` sets the softlayer default.
  - `--softlayer-domain`: **required** domain name for the machine
  - `--softlayer-hostname`: hostname for the machine
  - `--softlayer-hourly-billing`: Sets the hourly billing flag (default), otherwise uses monthly billing
  - `--softlayer-image`: OS Image to use
  - `--softlayer-install-script`: custom install script to use for installing Docker, other setup actions
  - `--softlayer-local-disk`: Use local machine disk instead of softlayer SAN.
  - `--softlayer-memory`: Memory for host in MB
  - `--softlayer-private-net-only`: Disable public networking
  - `--softlayer-region`: softlayer region

#### Microsoft Azure
Create machines on [Microsoft Azure](http://azure.microsoft.com/).

You need to create a subscription with a cert. Run these commands:

    $ openssl req -x509 -nodes -days 365 -newkey rsa:1024 -keyout mycert.pem -out mycert.pem
    $ openssl pkcs12 -export -out mycert.pfx -in mycert.pem -name "My Certificate"
    $ openssl x509 -inform pem -in mycert.pem -outform der -out mycert.cer

Go to the Azure portal, go to the "Settings" page, then "Manage Certificates" and upload `mycert.cer`.

Grab your subscription ID from the portal, then run `machine create` with these details:

    $ machine create -d azure --azure-subscription-id="SUB_ID" --azure-subscription-cert="mycert.pem"

Options:

 - `--azure-subscription-id`: Your Azure subscription ID.
 - `--azure-subscription-cert`: Your Azure subscription cert.

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
 - `--openstack-docker-install`: Boolean flag to indicate if docker have to be installed on the machine. Useful when
   docker is already installed and configured in the OpenStack image. Default set to `true`

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
| `OS_API_KEY`         | `--rackspace-ap-key`        |
| `OS_REGION_NAME`     | `--rackspace-region`        |
| `OS_ENDPOINT_TYPE`   | `--rackspace-endpoint-type` |

#### VirtualBox
Creates machines locally on [VirtualBox](https://www.virtualbox.org/). Requires VirtualBox to be installed.

Options:

 - `--virtualbox-boot2docker-url`: The URL of the boot2docker image. Defaults to the latest available version.
 - `--virtualbox-disk-size`: Size of disk for the host in MB. Default: `20000`
 - `--virtualbox-memory`: Size of memory for the host in MB. Default: `1024`

#### VMware Fusion
Creates machines locally on [VMware Fusion](http://www.vmware.com/products/fusion). Requires VMware Fusion to be installed.

Options:

 - `--vmwarefusion-boot2docker-url`: URL for boot2docker image.
 - `--vmwarefusion-disk-size`: Size of disk for host VM (in MB). Default: `20000`
 - `--vmwarefusion-memory-size`: Size of memory for host VM (in MB). Default: `1024`

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


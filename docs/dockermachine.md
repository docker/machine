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
machine version 0.0.2
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
INFO[0038] "dev" has been created and is now the active machine. To point Docker at this machine, run: export DOCKER_HOST=$(machine url) DOCKER_AUTH=identity
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
Docker to talk to that machine directly by setting the `DOCKER_HOST` 
and `DOCKER_AUTH` environment variables:

```
$ export DOCKER_HOST=$(machine url) DOCKER_AUTH=identity
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
INFO[0085] "staging" has been created and is now the active machine. To point Docker at this machine, run: export DOCKER_HOST=$(machine url) DOCKER_AUTH=identity
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
accept remote connections over TCP using 
[libtrust](https://github.com/docker/libtrust) for authentication.  Once this 
is finished, the host is ready for connection.

Just like with in the last section, we must run:

```
$ export DOCKER_HOST=$(machine url) DOCKER_AUTH=identity
```

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

To select an active host, you can use the `machine active` command.  You must 
re-run the `export` commands previously mentioned.

```
$ machine active dev
$ export DOCKER_HOST=$(machine url) DOCKER_AUTH=identity
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
$ export DOCKER_HOST=$(machine url) DOCKER_AUTH=identity
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

#### url

Get the URL of a host

```
$ machine url
tcp://192.168.99.109:2376
```

## Driver Options

TODO: List all possible values (where applicable) for all flags for every 
driver.

#### VirtualBox
#### Digital Ocean
#### Microsoft Azure
#### Google Compute Engine
#### Amazon Web Services
#### IBM Softlayer

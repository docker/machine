<!--[metadata]>
+++
title = "create"
description = "Create a machine."
keywords = ["machine, create, subcommand"]
[menu.main]
identifier="machine.create"
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# create

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

## Filtering create flags by driver in the help text

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
   --engine-env                                                                                         Specify environment variables to set in the engine
   --swarm                                                                                              Configure Machine with Swarm
   --swarm-master                                                                                       Configure Machine to be a Swarm master
   --swarm-discovery                                                                                    Discovery service to use with Swarm
   --swarm-host "tcp://0.0.0.0:3376"                                                                    ip/socket to listen on for Swarm master
   --swarm-addr                                                                                         addr to advertise for Swarm (default: detect and use the machine IP)
```

## Specifying configuration options for the created Docker engine

As part of the process of creation, Docker Machine installs Docker and
configures it with some sensible defaults. For instance, it allows connection
from the outside world over TCP with TLS-based encryption and defaults to AUFS
as the [storage
driver](https://docs.docker.com/reference/commandline/daemon/#daemon-storage-driver-option) when
available.

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
    --engine-storage-driver overlay \
    --engine-insecure-registry registry.myco.com \
    foobarmachine
```

This will create a virtual machine running locally in Virtualbox which uses the
`overlay` storage backend, has the key-value pairs `foo=bar` and `spam=eggs` as
labels on the engine, and allows pushing / pulling from the insecure registry
located at `registry.myco.com`. You can verify much of this by inspecting the
output of `docker info`:

```
$ eval $(docker-machine env foobarmachine)
$ docker version
Containers: 0
Images: 0
Storage Driver: overlay
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
- `--engine-registry-mirror`: Specify [registry mirrors](https://github.com/docker/distribution/blob/master/docs/mirror.md) to use
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

Additionally, Docker Machine supports a flag, `--engine-env`, which can be used to
specify arbitrary environment variables to be set within the engine with the syntax `--engine-env name=value`. For example, to specify that the engine should use `example.com` as the proxy server, you could run the following create command:

```
$ docker-machine create -d virtualbox \
    --engine-env HTTP_PROXY=http://example.com:8080 \
    --engine-env HTTPS_PROXY=https://example.com:8080 \
    --engine-env NO_PROXY=example2.com \
    proxbox
```

## Specifying Docker Swarm options for the created machine

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

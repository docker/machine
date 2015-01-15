# Docker Machine

Machine makes it really easy to create Docker hosts on local hypervisors and cloud providers. It creates servers, installs Docker on them, then configures the Docker client to talk to them.

It works a bit like this:

```console
$ machine create -d virtualbox dev
[info] Downloading boot2docker...
[info] Creating SSH key...
[info] Creating VirtualBox VM...
[info] Starting VirtualBox VM...
[info] Waiting for VM to start...
[info] "dev" has been created and is now the active host. Docker commands will now run against that host.

$ machine ls
NAME  	ACTIVE   DRIVER     	STATE 	URL
dev   	*    	virtualbox 	Running   tcp://192.168.99.100:2375

$ export DOCKER_HOST=`machine url` DOCKER_AUTH=identity

$ docker run busybox echo hello world
Unable to find image 'busybox' locally
Pulling repository busybox
e72ac664f4f0: Download complete
511136ea3c5a: Download complete
df7546f9f060: Download complete
e433a6c5b276: Download complete
hello world

$ machine create -d digitalocean --digitalocean-access-token=... staging
[info] Creating SSH key...
[info] Creating Digital Ocean droplet...
[info] Waiting for SSH...
[info] "staging" has been created and is now the active host. Docker commands will now run against that host.

$ machine ls
NAME      ACTIVE   DRIVER         STATE     URL
dev                virtualbox     Running   tcp://192.168.99.108:2376
staging   *        digitalocean   Running   tcp://104.236.37.134:2376
```

Machine creates Docker hosts that are secure by default. The connection between the client and daemon is encrypted and authenticated using new identity-based authentication. If you'd like to learn more about this, it is being worked on in [a pull request on Docker](https://github.com/docker/docker/pull/8265).

## Try it out

Machine is still in its early stages. If you'd like to try out a preview build, [download it here](https://github.com/docker/machine/releases/latest).

You will also need a version of Docker with identity authentication. Builds are available here:

 - Mac OS X: https://ejhazlett.s3.amazonaws.com/public/docker/darwin/docker-1.4.1-136b351e-identity
 - Linux: https://ejhazlett.s3.amazonaws.com/public/docker/linux/docker-1.4.1-136b351e-identity

## Drivers

### VirtualBox

Creates machines locally on [VirtualBox](https://www.virtualbox.org/). Requires VirtualBox to be installed.

Options:

 - `--virtualbox-boot2docker-url`: The URL of the boot2docker image. Defaults to the latest available version.
 - `--virtualbox-disk-size`: Size of disk for the host in MB. Default: `20000`
 - `--virtualbox-memory`: Size of memory for the host in MB. Default: `1024`

### Digital Ocean

Creates machines on [Digital Ocean](https://www.digitalocean.com/). You need to create a personal access token under "Apps & API" in the Digital Ocean Control Panel and pass that to `machine create` with the `--digitalocean-access-token` option.

Options:

 - `--digitalocean-access-token`: Your personal access token for the Digital Ocean API.
 - `--digitalocean-image`: The name of the Digital Ocean image to use. Default: `docker`
 - `--digitalocean-region`: The region to create the droplet in. Default: `nyc3`
 - `--digitalocean-size`: The size of the Digital Ocean driver. Default: `512mb`

### Microsoft Azure

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

### Amazon EC2

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

### Google Compute Engine

Create machines on [Google Compute Engine](https://cloud.google.com/compute/).  You will need a Google account and project name.  See https://cloud.google.com/compute/docs/projects for details on projects.

The Google driver uses oAuth.  When creating the machine, you will have your browser opened to authorize.  Once authorized, paste the code given in the prompt to launch the instance.

Options:

 - `--google-zone`: The zone to launch the instance.  Default: `us-central1-a`
 - `--google-machine-type`: The type of instance.  Default: `f1-micro`
 - `--google-username`: The username to use for the instance.  Default: `docker-user`
 - `--google-instance-name`: The name of the instance.  Default: `docker-machine`
 - `--google-project`: The name of your project to use when launching the instance.

### VMware Fusion

Creates machines locally on [VMware Fusion](http://www.vmware.com/products/fusion). Requires VMware Fusion to be installed.

Options:

 - `--vmwarefusion-boot2docker-url`: URL for boot2docker image.
 - `--vmwarefusion-disk-size`: Size of disk for host VM (in MB). Default: `20000`
 - `--vmwarefusion-memory-size`: Size of memory for host VM (in MB). Default: `1024`

### VMware vCloud Air

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

### VMware vSphere

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

### OpenStack

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

### Rackspace

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

## Contributing

[![GoDoc](https://godoc.org/github.com/docker/machine?status.png)](https://godoc.org/github.com/docker/machine)
[![Build Status](https://travis-ci.org/docker/machine.svg?branch=master)](https://travis-ci.org/docker/machine)

Want to hack on Machine? [Docker's contributions guidelines](https://github.com/docker/docker/blob/master/CONTRIBUTING.md) apply.

The requirements to build Machine are:

1. A running instance of Docker
2. The `bash` shell

To build, run:

    $ script/build

From the Machine repository's root.  Machine will run the build inside of a
Docker container and the compiled binaries will appear in the project directory
on the host.

By default, Machine will run a build which cross-compiles binaries for a variety
of architectures and operating systems.  If you know that you are only compiling
for a particular architecture and/or operating system, you can speed up
compilation by overriding the default argument that the build script passes
to [gox](https://github.com/mitchellh/gox).  This is very useful if you want
to iterate quickly on a new feature, bug fix, etc.

For instance, if you only want to compile for use on OSX with the x86_64 arch,
run:

    $ script/build -osarch="darwin/amd64"

If you have any questions we're in #docker-machine on Freenode.

=======
## Integration Tests
There is a suite of integration tests that will run for the drivers.  In order
to use these you must export the corresponding environment variables for each
driver as these perform the actual actions (start, stop, restart, kill, etc).

By default, the suite will run tests against all drivers in master.  You can
override this by setting the environment variable `MACHINE_TESTS`.  For example,
`MACHINE_TESTS="virtualbox" ./script/run-integration-tests` will only run the
virtualbox driver integration tests.

To run, use the helper script `./script/run-integration-tests`.

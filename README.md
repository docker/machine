# Docker Machine

Machine makes it really easy to create Docker hosts on local hypervisors and cloud providers. It creates servers, installs Docker on them, then configures the Docker client to talk to them.

It works a bit like this:

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

Machine creates Docker hosts that are secure by default. The connection between the client and daemon is encrypted and authenticated using new identity-based authentication. If you'd like to learn more about this, it is being worked on in [a pull request on Docker](https://github.com/docker/docker/pull/8265).

## Try it out

Machine is still in its early stages. If you'd like to try out a preview build, [download it here](https://github.com/docker/machine/releases/latest).

You will also need a version of Docker with identity authentication. Builds are available here:

 - Mac OS X: https://bfirsh.s3.amazonaws.com/docker/darwin/docker-1.3.1-dev-identity-auth
 - Linux: https://bfirsh.s3.amazonaws.com/docker/linux/docker-1.3.1-dev-identity-auth

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
 - `--amazonec2-vpc-id`: Your VPC ID to launch the instance in.
 - `--amazonec2-zone`: The AWS zone launch the instance in (i.e. one of a,b,c,d,e).

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

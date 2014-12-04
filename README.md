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

Machine is still in its early stages. If you'd like to try out a preview build, download it here:

 - Mac OS X: https://github.com/docker/machine/releases/download/0.0.1/darwin
 - Linux: https://github.com/docker/machine/releases/download/0.0.1/linux

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

    $ machine create -d azure --azure-subscription-id="SUB_ID" --azure-subscription-cert="mycert.cer"

Options:

 - `--azure-subscription-id`: Your Azure subscription ID.
 - `--azure-subscription-cert`: Your Azure subscription cert.

## Contributing

Want to hack on Machine? [Docker's contributions guidelines](https://github.com/docker/docker/blob/master/CONTRIBUTING.md) apply, and you'll also need [Docker with identity authentication](https://github.com/dmcgowan/docker/tree/tls_libtrust_auth) checked out in your GOPATH.

## Creators

**Ben Firshman**

- <https://twitter.com/bfirsh>
- <https://github.com/bfirsh>


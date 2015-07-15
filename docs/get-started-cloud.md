<!--[metadata]>
+++
title = "Using Docker Machine with a cloud provider"
description = "Using Docker Machine with a cloud provider"
keywords = ["docker, machine, amazonec2, azure, digitalocean, google, openstack, rackspace, softlayer, virtualbox, vmwarefusion, vmwarevcloudair, vmwarevsphere, exoscale"]
[menu.main]
parent="smn_workw_machine"
weight=2
+++
<![end-metadata]-->

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
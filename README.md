# ⚠️This is a fork of Docker Machine ⚠

This is a fork of Docker Machine maintained by GitLab for [fixing critical bugs](https://docs.gitlab.com/runner/executors/docker_machine.html#forked-version-of-docker-machine). Docker Machine, which Docker has deprecated as of 2021-09-27, is the basis of the GitLab Runner Docker Machine Executor. Our plan, as discussed [here](https://gitlab.com/gitlab-org/gitlab/-/issues/341856), is to continue to maintain the fork in the near term, with a primary focus on driver maintenance for Amazon Web Services, Google Cloud Platform, Microsoft Azure.

For a new merge request to be considered, the following questions must be answered: 

  * What critical bug this MR is fixing?
  * How does this change help reduce cost of usage? What scale of cost reduction is it?
  * In what scenarios is this change usable with GitLab Runner's `docker+machine` executor? 

Builds from this fork can be downloaded at https://gitlab-docker-machine-downloads.s3.amazonaws.com/main/index.html

# Docker Machine

![](https://docs.docker.com/machine/img/logo.png)

Machine lets you create Docker hosts on your computer, on cloud providers, and
inside your own data center. It creates servers, installs Docker on them, then
configures the Docker client to talk to them.

It works a bit like this:

```console
$ docker-machine create -d virtualbox default
Running pre-create checks...
Creating machine...
(default) Creating VirtualBox VM...
(default) Creating SSH key...
(default) Starting VM...
Waiting for machine to be running, this may take a few minutes...
Machine is running, waiting for SSH to be available...
Detecting operating system of created instance...
Detecting the provisioner...
Provisioning with boot2docker...
Copying certs to the local machine directory...
Copying certs to the remote machine...
Setting Docker configuration on the remote daemon...
Checking connection to Docker...
Docker is up and running!
To see how to connect Docker to this machine, run: docker-machine env default

$ docker-machine ls
NAME      ACTIVE   DRIVER       STATE     URL                         SWARM   DOCKER   ERRORS
default   -        virtualbox   Running   tcp://192.168.99.188:2376           v1.9.1

$ eval "$(docker-machine env default)"

$ docker run busybox echo hello world
Unable to find image 'busybox:latest' locally
511136ea3c5a: Pull complete
df7546f9f060: Pull complete
ea13149945cb: Pull complete
4986bf8c1536: Pull complete
hello world
```

In addition to local VMs, you can create and manage cloud servers:

```console
$ docker-machine create -d digitalocean --digitalocean-access-token=secret staging
Creating SSH key...
Creating Digital Ocean droplet...
To see how to connect Docker to this machine, run: docker-machine env staging

$ docker-machine ls
NAME      ACTIVE   DRIVER         STATE     URL                         SWARM   DOCKER   ERRORS
default   -        virtualbox     Running   tcp://192.168.99.188:2376           v1.9.1
staging   -        digitalocean   Running   tcp://203.0.113.81:2376             v1.9.1
```

## Installation and documentation

Full documentation [is available here](/docs/install-machine.md).

## Troubleshooting

Docker Machine tries to do the right thing in a variety of scenarios but
sometimes things do not go according to plan.  Here is a quick troubleshooting
guide which may help you to resolve of the issues you may be seeing.

Note that some of the suggested solutions are only available on the Docker
Machine default branch.  If you need them, consider compiling Docker Machine from
source.

#### `docker-machine` hangs

A common issue with Docker Machine is that it will hang when attempting to start
up the virtual machine.  Since starting the machine is part of the `create`
process, `create` is often where these types of errors show up.

A hang could be due to a variety of factors, but the most common suspect is
networking.  Consider the following:

-   Are you using a VPN?  If so, try disconnecting and see if creation will
    succeed without the VPN.  Some VPN software aggressively controls routes and
    you may need to [manually add the route](https://github.com/docker/machine/issues/1500#issuecomment-121134958).
-   Are you connected to a proxy server, corporate or otherwise?  If so, take a
    look at the `--no-proxy` flag for `env` and at [setting environment variables
    for the created Docker Engine](https://docs.docker.com/machine/reference/create/#specifying-configuration-options-for-the-created-docker-engine).
-   Are there a lot of host-only interfaces listed by the command `VBoxManage list
    hostonlyifs`?  If so, this has sometimes been known to cause bugs.  Consider
    removing the ones you are not using (`VBoxManage hostonlyif remove name`) and
    trying machine creation again.


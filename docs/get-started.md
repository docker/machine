<!--[metadata]>
+++
title = "Get started with Machine and a local VM"
description = "Get started with Docker Machine and a local VM"
keywords = ["docker, machine, virtualbox, local"]
[menu.main]
parent="workw_machine"
weight=1
+++
<![end-metadata]-->

# Get started with Docker Machine and a local VM

Let's take a look at using `docker-machine` for creating, using, and managing a
Docker host inside of [VirtualBox](https://www.virtualbox.org/).

First, ensure that [the latest
VirtualBox](https://www.virtualbox.org/wiki/Downloads) is correctly installed
on your system.

If you run the `docker-machine ls` command to show all available machines, you
will see that none have been created so far.

    $ docker-machine ls
    NAME   ACTIVE   DRIVER   STATE   URL   SWARM   DOCKER   ERRORS

To create one, we run the `docker-machine create` command, passing the string
`virtualbox` to the `--driver` flag. The final argument we pass is the name of
the machine - in this case, we will name our machine "default".

This command will download a lightweight Linux distribution
([boot2docker](https://github.com/boot2docker/boot2docker)) with the Docker
daemon installed, and will create and start a VirtualBox VM with Docker
running.

    $ docker-machine create --driver virtualbox default
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

You can see the machine you have created by running the `docker-machine ls`
command again:

    $ docker-machine ls
    NAME      ACTIVE   DRIVER       STATE     URL                         SWARM   DOCKER   ERRORS
    default   *        virtualbox   Running   tcp://192.168.99.187:2376           v1.9.1

Next, as noted in the output of the `docker-machine create` command, we have to
tell Docker to talk to that machine. You can do this with the `docker-machine
env` command. For example,

    $ eval "$(docker-machine env default)"
    $ docker ps

> **Note**: If you are using `fish`, or a Windows shell such as
> Powershell/`cmd.exe` the above method will not work as described. Instead,
> see [the `env` command's documentation](reference/env.md)
> to learn how to set the environment variables for your shell.

This will set environment variables that the Docker client will read which
specify the TLS settings. Note that you will need to do that every time you
open a new tab or restart your machine.

To see what will be set, we can run `docker-machine env default`.

    $ docker-machine env default
    export DOCKER_TLS_VERIFY="1"
    export DOCKER_HOST="tcp://172.16.62.130:2376"
    export DOCKER_CERT_PATH="/Users/<yourusername>/.docker/machine/machines/default"
    export DOCKER_MACHINE_NAME="default"
    # Run this command to configure your shell:
    # eval "$(docker-machine env default)"

You can now run Docker commands on this host:

    $ docker run busybox echo hello world
    Unable to find image 'busybox' locally
    Pulling repository busybox
    e72ac664f4f0: Download complete
    511136ea3c5a: Download complete
    df7546f9f060: Download complete
    e433a6c5b276: Download complete
    hello world

Any exposed ports are available on the Docker host’s IP address, which you can
get using the `docker-machine ip` command:

    $ docker-machine ip default
    192.168.99.100

For instance, you can try running a webserver ([nginx](https://www.nginx.com/)
in a container with the following command:

    $ docker run -d -p 8000:80 nginx

When the image is finished pulling, you can hit the server at port 8000 on the
IP address given to you by `docker-machine ip`. For instance:

    $ curl $(docker-machine ip default):8000
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

You can create and manage as many local VMs running Docker as you please- just
run `docker-machine create` again. All created machines will appear in the
output of `docker-machine ls`.

If you are finished using a host for the time being, you can stop it with
`docker-machine stop` and later start it again with `docker-machine start`.

    $ docker-machine stop default
    $ docker-machine start default

## Operating on machines without specifying the name

Some commands will assume that the specified operation should be run on the default
machine (if it exists) if no machine name is passed to them as an argument. The default
machine name can be specified in an environment variable `DOCKER_MACHINE_NAME`, or is
simply named `default` if this environment variable does not exist. This allows you to
save some typing on Machine commands that may be frequently invoked.

For instance:

    $ docker-machine stop
    Stopping "default"....
    Machine "default" was stopped.

    $ docker-machine start
    Starting "default"...
    (default) Waiting for an IP...
    Machine "default" was started.
    Started machines may have new IP addresses.  You may need to re-run the `docker-machine env` command.

    $ eval $(docker-machine env)

    $ docker-machine ip
    192.168.99.100

And, because `docker-machine env` also defines `DOCKER_MACHINE_NAME`:

    $ eval $(docker-machine env staging)

    $ docker-machine start
    Starting "staging"...
    Machine "staging" is already running.

    $ docker-machine ip
    203.0.113.81

Commands that will follow this style are:

- `docker-machine config`
- `docker-machine env`
- `docker-machine inspect`
- `docker-machine ip`
- `docker-machine kill`
- `docker-machine provision`
- `docker-machine regenerate-certs`
- `docker-machine restart`
- `docker-machine ssh`
- `docker-machine start`
- `docker-machine status`
- `docker-machine stop`
- `docker-machine upgrade`
- `docker-machine url`

For machines other than the default machine, and commands other than those listed above,
you must always specify the name explicitly as an argument.

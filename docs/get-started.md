<!--[metadata]>
+++
title = "Get started with Docker Machine and a local VM"
description = "Get started with Docker Machine and a local VM"
keywords = ["docker, machine, virtualbox, local"]
[menu.main]
parent="smn_workw_machine"
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

    $ eval "$(docker-machine env dev)"
    $ docker ps

> **Note**: If you are using `fish`, or a Windows shell such as
> Powershell/`cmd.exe` the above method will not work as described. Instead,
> see [the `env` command's documentation](reference/env.md)
> to learn how to set the environment variables for your shell.

This will set environment variables that the Docker client will read which
specify the TLS settings. Note that you will need to do that every time you
open a new tab or restart your machine.

To see what will be set, run `docker-machine env dev`.

    $ docker-machine env dev
    export DOCKER_TLS_VERIFY="1"
    export DOCKER_HOST="tcp://172.16.62.130:2376"
    export DOCKER_CERT_PATH="/Users/<your username>/.docker/machine/machines/dev"
    export DOCKER_MACHINE_NAME="dev"
    # Run this command to configure your shell:
    # eval "$(docker-machine env dev)"

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

    $ docker-machine ip dev
    192.168.99.100

For instance, you can try running a webserver ([nginx](https://www.nginx.com/)
in a container with the following command:

    $ docker run -d -p 8000:80 nginx

When the image is finished pulling, you can hit the server at port 8000 on the
IP address given to you by `docker-machine ip`. For instance:

    $ curl $(docker-machine ip dev):8000
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
Make sure to specify the machine name as an argument:

    $ docker-machine stop dev
    $ docker-machine start dev

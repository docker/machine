<!--[metadata]>
+++
title = "config"
description = "Show client configuration"
keywords = ["machine, config, subcommand"]
[menu.main]
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# config

Show the Docker client configuration for a machine.

    $ docker-machine config dev
    --tlsverify
    --tlscacert="/home/bjaglin/.docker/machine/certs/ca.pem"
    --tlscert="/home/bjaglin/.docker/machine/certs/cert.pem"
    --tlskey="/home/bjaglin/.docker/machine/certs/key.pem"
    -H=tcp://192.168.99.100:2376

You can pass the arguments right after the `docker` command via command
substitution on Unix shells:

    $ docker $(docker-machine config dev) info
    Containers: 58
    Images: 545
    Server Version: 1.9.1
    Storage Driver: overlay
     Backing Filesystem: extfs
    Execution Driver: native-0.2
    Logging Driver: json-file
    Kernel Version: 4.2.0-19-generic
    Operating System: Ubuntu 15.10
    CPUs: 4
    Total Memory: 11.64 GiB
    Name: t440s
    ID: 4CO2:NGIJ:KC6L:MBYG:BXNH:DRJK:GM5J:LQHT:H6QD:7RA2:FX2O:NRXF
    WARNING: No swap limit support


## Docker in Docker

If you use want to forward the configuration to a Docker client running within
a container, you can start this container, use the `--dind` flag:

    $ docker-machine config --dind b2d
    --volume="/home/bjaglin/.docker/machine/machines/b2d:/etc/docker/cert:ro"
    --env="DOCKER_CERT_PATH=/etc/docker/cert"
    --env="DOCKER_TLS_VERIFY=1"
    --env="DOCKER_HOST=tcp://192.168.99.100:2376"

Note that the configuration should passed as arguments to the `docker run`
command and not to the `docker` client itself, as the container being started
will be run on the local engine (or a remote one if `DOCKER_HOST` has been set
earlier).

    $ docker run $(docker-machine config --dind b2d) docker:1.9.0 docker version
    Client:
     Version:      1.9.0
     API version:  1.21
     Go version:   go1.4.3
     Git commit:   76d6bc9
     Built:        Tue Nov  3 19:20:09 UTC 2015
     OS/Arch:      linux/amd64
    
    Server:
     Version:      1.9.1
     API version:  1.21
     Go version:   go1.4.3
     Git commit:   a34a1d5
     Built:        Fri Nov 20 17:56:04 UTC 2015
     OS/Arch:      linux/amd64

This is an advanced use-case, useful when using containers relying on a Docker
client available within the container, typically for orchestration. The Docker
client within the container might be part of the image, or bind-mounted (which
requires the binary to be statically linked).

<!--[metadata]>
+++
title = "Docker Machine"
description = "How to install Docker Machine"
keywords = ["machine, orchestration, install, installation, docker, documentation"]
[menu.main]
parent="mn_install"
weight=3
+++
<![end-metadata]-->

# Install Docker Machine

Docker Machine is supported on Windows, OS X, and Linux and is installable as
one standalone binary.  The links to the binaries for the various platforms and
architectures are available at the [Github
Release](https://github.com/docker/machine/releases/) page.


## OS X and Windows

Install Machine using the Docker Toolbox using the <a href="https://docs.docker.com/installation/mac/" target="_blank">Mac OS X installation</a>
instruction or <a href="https://docs.docker.com/installation/windows" target="_blank">Windows installation</a> instructions.

## On Linux

To install on Linux, do the following:

1. Install <a href="https://docs.docker.com/installation/" target="_blank">Docker version 1.7.1 or greater</a>:

2. Download the Machine binary to somewhere in your `PATH` (for example, `/usr/local/bin`).

        $ curl -L https://github.com/docker/machine/releases/download/v0.4.0/docker-machine_linux-amd64 > /usr/local/bin/docker-machine

3. Apply executable permissions to the binary:

        $ chmod +x /usr/local/bin/docker-machine

4. Check the installation by displaying the Machine version:

			$ docker-machine -v
			machine version 0.4.0

## Where to go next

* [Docker Machine overview](/)
* [Docker Machine driver reference](/drivers)
* [Docker Machine subcommand reference](/reference)


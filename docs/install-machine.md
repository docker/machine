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

Docker Machine is supported on Windows, OS X, and Linux operating systems. You
can install using one of Docker's automated installation methods or you can
download and install via a binary. This page details each of those methods.

## OS X and Windows

On OS X and Windows, Machine is installed along with other Docker products when
you install the Docker Toolbox. For details on installing Docker Toolbox, see
the <a href="https://docs.docker.com/installation/mac/" target="_blank">Mac OS X
installation</a> instructions or <a
href="https://docs.docker.com/installation/windows" target="_blank">Windows
installation</a> instructions.

If you only want Docker Machine, you can install [the Machine binaries
directly](https://github.com/docker/machine/releases/). Alternatively, OS X
users have the option to follow the Linux installation instructions.

## On Linux

To install on Linux, do the following:

1. Install <a href="https://docs.docker.com/installation/"
target="_blank">Docker version 1.7.1 or greater</a>:

2. Download the Machine binary to somewhere in your `PATH` (for example,
`/usr/local/bin`).

        $ curl -L https://github.com/docker/machine/releases/download/v0.4.0/docker-machine_linux-amd64 > /usr/local/bin/docker-machine

3. Apply executable permissions to the binary:

        $ chmod +x /usr/local/bin/docker-machine

4. Check the installation by displaying the Machine version:

			$ docker-machine -v
			machine version 0.4.0

## Install from binary

The Docker Machine team compiles binaries for several platforms and
architectures and makes them available from [the Machine release page on
Github](https://github.com/docker/machine/releases/). To install from a binary:

1. Download the binary you want.
2. Rename the binary to `docker-machine`.
3. Move the `docker-machine` file to an appropriate directory on your system.

    For example, on an OS X machine, you might move it to the `/usr/local/bin`
    directory.

4. Ensure the file's executable permissions are correct.
5. Apply executable permissions to the binary:

        $ chmod +x /usr/local/bin/docker-machine

6. Check the installation by displaying the Machine version:

        $ docker-machine -v
        machine version 0.4.0

## Where to go next

* [Docker Machine overview](/)
* [Docker Machine driver reference](/drivers)
* [Docker Machine subcommand reference](/reference)

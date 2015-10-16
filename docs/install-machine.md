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

On OS X and Windows, Machine is installed along with other Docker products when
you install the Docker Toolbox. For details on installing Docker Toolbox, see
the <a href="https://docs.docker.com/installation/mac/" target="_blank">Mac OS X
installation</a> instructions or <a
href="https://docs.docker.com/installation/windows" target="_blank">Windows
installation</a> instructions.

If you only want Docker Machine, you can install the Machine binaries (the
latest versions of which are located at
https://github.com/docker/machine/releases/) directly by following the
instructions in the next section.

## Installing Machine Directly

1. Install <a href="https://docs.docker.com/installation/"
target="_blank">the Docker binary</a>.

2. Download the archive containing the Docker Machine binaries and extract them
to your PATH.

Linux:

```
$ curl -L https://github.com/docker/machine/releases/download/v0.5.0/docker-machine_linux-amd64.zip >machine.zip && \
  unzip machine.zip && \
  rm machine.zip && \
  mv docker-machine* /usr/local/bin
```

OSX:

```
$ curl -L https://github.com/docker/machine/releases/download/v0.5.0/docker-machine_darwin-amd64.zip >machine.zip && \
  unzip machine.zip && \
  rm machine.zip && \
  mv docker-machine* /usr/local/bin
```

Windows (using Git Bash):

```
$ curl -L https://github.com/docker/machine/releases/download/v0.5.0/docker-machine_windows-amd64.zip >machine.zip && \
  unzip machine.zip && \
  rm machine.zip && \
  mv docker-machine* /usr/local/bin
```

3. Check the installation by displaying the Machine version:

```
$ docker-machine -v
machine version 0.5.0 (3e06852)
```

## Where to go next

* [Docker Machine overview](/)
* [Docker Machine driver reference](drivers/index.md)
* [Docker Machine subcommand reference](reference/index.md)

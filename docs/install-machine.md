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

## Install Docker Machine

Docker Machine is supported on Windows, OS X, and Linux and is installable as
one standalone binary.  The links to the binaries for the various platforms and
architectures are available at the [Github
Release](https://github.com/docker/machine/releases/) page.


### OS X and Linux

To install on OS X or Linux, download the proper binary to somewhere in your
`PATH` (e.g. `/usr/local/bin`) and make it executable. For instance, to install on
most OS X machines these commands should suffice:

```
$ curl -L https://github.com/docker/machine/releases/download/v0.3.1/docker-machine_darwin-amd64 > /usr/local/bin/docker-machine
$ chmod +x /usr/local/bin/docker-machine
```

For Linux, just substitute "linux" for "darwin" in the binary name above.

Now you should be able to check the version with `docker-machine -v`:

```
$ docker-machine -v
machine version 0.3.1
```

In order to run Docker commands on your machines without having to use SSH, make
sure to install the Docker client as well, e.g.:

```
$ curl -L https://get.docker.com/builds/Darwin/x86_64/docker-latest > /usr/local/bin/docker
$ chmod +x /usr/local/bin/docker
```

### Windows

Currently, Docker recommends that you install and use Docker Machine on Windows
with [msysgit](https://msysgit.github.io/). This will provide you with some
programs that Docker Machine relies on such as `ssh`, as well as a functioning
shell.

When you have installed msysgit, start up the terminal prompt and run the
following commands. Here it is assumed that you are on a 64-bit Windows
installation. If you are on a 32-bit installation, please substitute "i386" for
"x86_64" in the URLs mentioned.

First, install the Docker client binary:

```
$ curl -L https://get.docker.com/builds/Windows/x86_64/docker-latest.exe > /bin/docker
```

Next, install the Docker Machine binary:

```
$ curl -L https://github.com/docker/machine/releases/download/v0.3.1/docker-machine_windows-amd64.exe > /bin/docker-machine
```

Now running `docker-machine` should work.

```
$ docker-machine -v
machine version 0.3.1
```

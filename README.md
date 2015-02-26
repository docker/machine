# Docker Machine

Machine makes it really easy to create Docker hosts on your computer, on cloud providers and inside your own data center. It creates servers, installs Docker on them, then configures the Docker client to talk to them.

It works a bit like this:

```console
$ docker-machine create -d virtualbox dev
[info] Downloading boot2docker...
[info] Creating SSH key...
[info] Creating VirtualBox VM...
[info] Starting VirtualBox VM...
[info] Waiting for VM to start...
[info] "dev" has been created and is now the active host. Docker commands will now run against that host.

$ docker-machine ls
NAME  	ACTIVE   DRIVER     	STATE 	URL
dev   	*    	virtualbox 	Running   tcp://192.168.99.100:2375

$ docker $(docker-machine config dev) run busybox echo hello world
Unable to find image 'busybox' locally
Pulling repository busybox
e72ac664f4f0: Download complete
511136ea3c5a: Download complete
df7546f9f060: Download complete
e433a6c5b276: Download complete
hello world

$ docker-machine create -d digitalocean --digitalocean-access-token=... staging
[info] Creating SSH key...
[info] Creating Digital Ocean droplet...
[info] Waiting for SSH...
[info] "staging" has been created and is now the active host. Docker commands will now run against that host.

$ docker-machine ls
NAME      ACTIVE   DRIVER         STATE     URL
dev                virtualbox     Running   tcp://192.168.99.108:2376
staging   *        digitalocean   Running   tcp://104.236.37.134:2376
```

## Installation and documentation

Full documentation [is available here](https://docs.docker.com/machine/).

## Contributing

[![GoDoc](https://godoc.org/github.com/docker/machine?status.png)](https://godoc.org/github.com/docker/machine)
[![Build Status](https://travis-ci.org/docker/machine.svg?branch=master)](https://travis-ci.org/docker/machine)

Want to hack on Machine? [Docker's contributions guidelines](https://github.com/docker/docker/blob/master/CONTRIBUTING.md) apply.

The requirements to build Machine are:

1. A running instance of Docker
2. The `bash` shell

To build, run:

    $ script/build

From the Machine repository's root.  Machine will run the build inside of a
Docker container and the compiled binaries will appear in the project directory
on the host.

By default, Machine will run a build which cross-compiles binaries for a variety
of architectures and operating systems.  If you know that you are only compiling
for a particular architecture and/or operating system, you can speed up
compilation by overriding the default argument that the build script passes
to [gox](https://github.com/mitchellh/gox).  This is very useful if you want
to iterate quickly on a new feature, bug fix, etc.

For instance, if you only want to compile for use on OSX with the x86_64 arch,
run:

    $ script/build -osarch="darwin/amd64"

If you have any questions we're in #docker-machine on Freenode.

## Unit Tests

To run the unit tests for the whole project, using the following script:

    $ script/test

This will run the unit tests inside of a container, so you don't have to worry
about configuring your environment properly before doing so.

To run the unit tests for only a specific subdirectory of the project, you can
pass an argument to that script to specify which directory, e.g.:

    $ script/test ./drivers/amazonec2

If you make a pull request, it is highly encouraged that you submit tests for
the code that you have added or modified in the same pull request.

## Code Coverage

Machine includes a script to check for missing `*_test.go` files and to generate
an [HTML-based repesentation of which code is covered by tests](http://blog.golang.org/cover#TOC_5.).

To run the code coverage script, execute:

```console
$ ./script/coverage serve
```

You will see the results of the code coverage check as they come in.

This will also generate the code coverage website and serve it from a container
on port 8000.  By default, `/` will show you the source files from the base
directory, and you can navigate to the coverage for any particular subdirectory
of the Docker Machine repo's root by going to that path.  For instance, to see
the coverage for the VirtualBox driver's package, browse to `/drivers/virtualbox`.

![](/docs/img/coverage.png)

You can hit `CTRL+C` to stop the server.

## Integration Tests
There is a suite of integration tests that will run for the drivers.  In order
to use these you must export the corresponding environment variables for each
driver as these perform the actual actions (start, stop, restart, kill, etc).

By default, the suite will run tests against all drivers in master.  You can
override this by setting the environment variable `MACHINE_TESTS`.  For example,
`MACHINE_TESTS="virtualbox" ./script/run-integration-tests` will only run the
virtualbox driver integration tests.

You can set the path to the machine binary under test using the `MACHINE_BINARY`
environment variable.

To run, use the helper script `./script/run-integration-tests`.

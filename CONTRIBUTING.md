# Contributing to machine

[![GoDoc](https://godoc.org/github.com/docker/machine?status.png)](https://godoc.org/github.com/docker/machine)
[![Build Status](https://travis-ci.org/docker/machine.svg?branch=master)](https://travis-ci.org/docker/machine)

Want to hack on Machine? Awesome! Here are instructions to get you
started.

Machine is a part of the [Docker](https://www.docker.com) project, and follows
the same rules and principles. If you're already familiar with the way
Docker does things, you'll feel right at home.

Otherwise, go read
[Docker's contributions guidelines](https://github.com/docker/docker/blob/master/CONTRIBUTING.md).

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
We utilize [BATS](https://github.com/sstephenson/bats) for integration testing.
This runs tests against the generated binary.  To use, make sure to install
BATS (use that link).  Then run `./script/build` to generate the binary.  Once
you have the binary, you can run test against a specified driver:

```
$ bats test/integration/driver-virtualbox.bats
 ✓ virtualbox: machine should not exist
 ✓ virtualbox: VM should not exist
 ✓ virtualbox: create
 ✓ virtualbox: active
 ✓ virtualbox: ls
 ✓ virtualbox: run busybox container 
 ✓ virtualbox: url
 ✓ virtualbox: ip
 ✓ virtualbox: ssh
 ✓ virtualbox: stop
 ✓ virtualbox: machine should show stopped
 ✓ virtualbox: start
 ✓ virtualbox: machine should show running after start
 ✓ virtualbox: restart
 ✓ virtualbox: machine should show running after restart
 ✓ virtualbox: remove
 ✓ virtualbox: machine should not exist
 ✓ virtualbox: VM should not exist

15 tests, 0 failures
```

You can also run the general `cli` tests:

```
$ bats test/integration/cli.bats
 ✓ cli: show info
 ✓ cli: show active help
 ✓ cli: show config help
 ✓ cli: show inspect help
 ✓ cli: show ip help
 ✓ cli: show kill help
 ✓ cli: show ls help
 ✓ cli: show restart help
 ✓ cli: show rm help
 ✓ cli: show env help
 ✓ cli: show ssh help
 ✓ cli: show start help
 ✓ cli: show stop help
 ✓ cli: show upgrade help
 ✓ cli: show url help
 ✓ flag: show version
 ✓ flag: show help

17 tests, 0 failures
```
# Drivers
Docker Machine has several included drivers that supports provisioning hosts
in various providers.  If you wish to contribute a driver, we ask the following
to ensure we keep the driver in a consistent and stable state:

- Address issues filed against this driver in a timely manner
- Review PRs for the driver
- Be responsible for maintaining the infrastructure to run unit tests
and integration tests on the new supported environment
- Participate in a weekly driver maintainer meeting

If you can commit to those, the next step is to make sure the driver adheres
to the [spec](https://github.com/docker/machine/blob/master/docs/DRIVER_SPEC.md).

Once you have created and tested the driver, you can open a PR.

Note: even if those are met does not guarantee a driver will be accepted.
If you have questions, please do not hesitate to contact us on IRC.

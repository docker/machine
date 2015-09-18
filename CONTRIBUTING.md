# Contributing to machine

[![GoDoc](https://godoc.org/github.com/docker/machine?status.png)](https://godoc.org/github.com/docker/machine)
[![Build Status](https://travis-ci.org/docker/machine.svg?branch=master)](https://travis-ci.org/docker/machine)
[![Coverage Status](https://coveralls.io/repos/docker/machine/badge.svg?branch=upstream-master&service=github)](https://coveralls.io/github/docker/machine?branch=upstream-master)

Want to hack on Machine? Awesome! Here are instructions to get you
started.

Machine is a part of the [Docker](https://www.docker.com) project, and follows
the same rules and principles. If you're already familiar with the way
Docker does things, you'll feel right at home.

Otherwise, please read [Docker's contributions
guidelines](https://github.com/docker/docker/blob/master/CONTRIBUTING.md).

# Building using docker

The requirements to build Machine are:

1. A running instance of Docker (or alternatively a golang 1.5 development environment)
2. The `bash` shell
3. [Make](https://www.gnu.org/software/make/)

Call `export USE_CONTAINER=true` to instruct the build system to use containers to build.
If you want to build natively using golang instead, don't set this variable.

## Building

To build the docker-machine binary, simply run:

    $ make

From the Machine repository's root. You will now find a `docker-machine`
binary at the root of the project.

You may call:

    $ make clean

to clean-up build results.

## Tests and validation

To run basic validation (dco, fmt), and the project unit tests, call:

    $ make test

If you want more indepth validation (vet, lint), and all tests with race detection, call:

    $ make validate

If you make a pull request, it is highly encouraged that you submit tests for
the code that you have added or modified in the same pull request.

## Code Coverage

To generate an html code coverage report of the Machine codebase, run:

    make coverage-serve

And navigate to http://localhost:8000 (hit `CTRL+C` to stop the server).

Alternatively, if you are building natively, you can simply run:

    make coverage-html

This will generate and open the report file:

![](/docs/img/coverage.png)

## List of all targets

### High-level targets

    make clean
    make build
    make test
    make validate

### Build targets

Build a single, native machine binary:

    make build-simple

Build for all supported oses and architectures (binaries will be in the `bin` project subfolder):

    make build-x

Build for a specific list of oses and architectures:

    TARGET_OS=linux TARGET_ARCH="amd64 arm" make build-x

You can further control build options through the following environment variables:

    DEBUG=true # enable debug build
    STATIC=true # build static (note: when cross-compiling, the build is always static)
    VERBOSE=true # verbose output
    PARALLEL=X # lets you control build parallelism when cross-compiling multiple builds
    PREFIX=folder

Scrub build results:

    make build-clean

### Coverage targets

    make coverage-html
    make coverage-serve
    make coverage-send
    make coverage-generate
    make coverage-clean

### Tests targets

    make test-short
    make test-long
    make test-integration

### Validation targets

    make fmt
    make vet
    make lint
    make dco

## Integration Tests

### Setup

We use [BATS](https://github.com/sstephenson/bats) for integration testing, so,
first make sure to [install it](https://github.com/sstephenson/bats#installing-bats-from-source).

### Basic Usage

Integration tests can be invoked calling `make test-integration`.

:warn: you cannot run integration test inside a container for now.
Be sure to unset the `USE_CONTAINER` env variable if you set it earlier, or alternatively
call directly `./test/integration/run-bats.sh` instead of `make test-integration`.

You can invoke a test or subset of tests for a particular driver.
To set the driver, use the `DRIVER` environment variable.

To invoke just one test:

```console
$ DRIVER=virtualbox make test-integration test/integration/core/core-commands.bats
 ✓ virtualbox: machine should not exist
 ✓ virtualbox: create
 ✓ virtualbox: ls
 ✓ virtualbox: run busybox container
 ✓ virtualbox: url
 ✓ virtualbox: ip
 ✓ virtualbox: ssh
 ✓ virtualbox: docker commands with the socket should work
 ✓ virtualbox: stop
 ✓ virtualbox: machine should show stopped after stop
 ✓ virtualbox: machine should now allow upgrade when stopped
 ✓ virtualbox: start
 ✓ virtualbox: machine should show running after start
 ✓ virtualbox: kill
 ✓ virtualbox: machine should show stopped after kill
 ✓ virtualbox: restart
 ✓ virtualbox: machine should show running after restart

17 tests, 0 failures
Cleaning up machines...
Successfully removed bats-virtualbox-test
```

To invoke a shared test with a different driver:

```console
$ DRIVER=digitalocean make test-integration test/integration/core/core-commands.bats
...
```

To invoke a directory of tests recursively:

```console
$ DRIVER=virtualbox make test-integration test/integration/core/
...
```

### Extra Create Arguments

In some cases, for instance to test the creation of a specific base OS (e.g.
RHEL) as opposed to the default with the common tests, you may want to run
common tests with different create arguments than you get out of the box.

Keep in mind that Machine supports environment variables for many of these
flags.  So, for instance, you could run the command (substituting, of course,
the proper secrets):

```
$ DRIVER=amazonec2 \
  AWS_VPC_ID=vpc-xxxxxxx \
  AWS_SECRET_ACCESS_KEY=yyyyyyyyyyyyy \
  AWS_ACCESS_KEY_ID=zzzzzzzzzzzzzzzz \
  AWS_AMI=ami-12663b7a \
  AWS_SSH_USER=ec2-user \
  make test-integration test/integration/core
```

in order to run the core tests on Red Hat Enterprise Linux on Amazon.

### Layout

The `test/integration` directory is layed out to divide up tests based on the
areas which the test.  If you are uncertain where to put yours, we are happy to
guide you.

At the time of writing, there is:

1. A `core` directory which contains tests that are applicable to all drivers.
2. A `drivers` directory which contains tests that are applicable only to
specific drivers with sub-directories for each provider.
3. A `cli` directory which is meant for testing functionality of the command
line interface, without much regard for driver-specific details.

### Guidelines

The best practices for writing integration tests on Docker Machine are still a
work in progress, but here are some general guidelines from the maintainers:

1.  Ideally, each test file should have only one concern.
2.  Tests generally should not spin up more than one machine unless the test is
deliberately testing something which involves multiple machines, such as an `ls`
test which involves several machines, or a test intended to create and check
some property of a Swarm cluster.
3.  BATS will print the output of commands executed during a test if the test
fails.  This can be useful, for instance to dump the magic `$output` variable
that BATS provides and/or to get debugging information.
4.  It is not strictly needed to clean up the machines as part of the test.  The
BATS wrapper script has a hook to take care of cleaning up all created machines
after each test.

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

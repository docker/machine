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

For instance, if you only want to compile for use on OS X with the x86_64 arch,
run:

    $ script/build -osarch="darwin/amd64"

If you don't need to run the `docker build` to generate the image on each
compile, i.e. if you have built the image already, you can skip the image build
using the `SKIP_BUILD` environment variable, for instance:

    $ SKIP_BUILD=1 script/build -osarch="darwin/amd64"

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
an [HTML-based representation of which code is covered by tests](http://blog.golang.org/cover#TOC_5.).

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

### Setup

We utilize [BATS](https://github.com/sstephenson/bats) for integration testing.
This runs tests against the generated binary.  To use, first make sure to
[install BATS](https://github.com/sstephenson/bats).  Then run `./script/build`
to generate the binary for your system.

### Basic Usage

Once you have the binary, the integration tests can be invoked using the
`test/integration/run-bats.sh` wrapper script.

Using this wrapper script, you can invoke a test or subset of tests for a
particular driver.  To set the driver, use the `DRIVER` environment variable.

The following examples are all shown relative to the project's root directory,
but you should be able to invoke them from any directory without issue.

To invoke just one test:

```console
$ DRIVER=virtualbox ./test/integration/run-bats.sh test/integration/core/core-commands.bats
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
$ DRIVER=digitalocean ./test/integration/run-bats.sh test/integration/core/core-commands.bats
...
```

To invoke a directory of tests recursively:

```console
$ DRIVER=virtualbox ./test/integration/run-bats.sh test/integration/core/
...
```

If you want to invoke a group of tests across two or more different drivers at
once (e.g. every test in the `drivers` directory), at the time of writing there
is no first-class support to do so - you will have to write your own wrapper
scripts, bash loops, etc.  However, in the future, this may gain first-class
support as usage patterns become more clear.

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
  ./test/integration/run-bats.sh test/integration/core
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

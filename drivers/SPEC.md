> DRAFT

# Machine Driver Specification v1
This is the standard configuration and specification for version 1 drivers.

Along with defining how a driver should provision instances, the standard
also discusses behavior and operations Machine expects.

# Requirements
The following are required for a driver to be included as a supported driver
for Docker Machine.

## Base Operating System
The provider must offer a base operating system supported by the Docker Engine.

## API Access
We prefer accessing the provider service via HTTP APIs and strongly recommend
using those over external executables.  For example, using the Amazon EC2 API
instead of the EC2 command line tools.  If in doubt, contact a project
maintainer.

## SSH
The provider must offer SSH access to control the instance.  This does not
have to be public, but must offer it as Machine relies on SSH for system
level maintenance.

## Maintainer
To be supported as an official driver, it will need to be maintained.  There
can be multiple driver maintainers and they will be identified in the 
maintainers file.

# Provider Operations
The following instance operations should be supported by the provider.

## Create
`Create` will launch a new instance and make sure it is ready for provisioning.
This includes setting up the instance with the proper SSH keys and making
sure SSH is available including any access control (firewall).  This should
return an error on failure.

## Remove
`Remove` will remove the instance from the provider.  This should remove the
instance and any associated services or artifacts that were created as part
of the instance including keys and access groups.  This should return an
error on failure.

## Start
`Start` will start a stopped instance.  This should ensure the instance is
ready for operations such as SSH and Docker.  This should return an error on
failure.

## Stop
`Stop` will stop a running instance.  This should ensure the instance is
stopped and return an error on failure.

## Kill
`Kill` will forcibly stop a running instance.  This should ensure the instance
is stopped and return an error on failure.

## Restart
`Restart` will restart a running instance.  This should ensure the instance
is ready for operations such as SSH and Docker.  This should return an error
on failure.

## Status
`Status` will return the state of the instance.  This should return the
current state of the instance (running, stopped, error, etc).  This should
return an error on failure.

# Testing
Testing is strongly recommended for drivers.  Unit tests are preferred as well
as inclusion into the [integration tests](https://github.com/docker/machine#integration-tests).

# Maintaining
Driver contributors are strongly encouraged to maintain the driver to keep
it supported.  We recommend and encourage contributors to join in the weekly
meetings to give feedback and participate in the development around Machine.
Driver maintainers will be notified and consulted for issues regarding their
driver.

# Third Party Libraries
If you want to use a third party library to interact with the provider, you
will need to make sure it is compliant with the Docker license terms (non-GPL).
For more information, contact a project maintainer.


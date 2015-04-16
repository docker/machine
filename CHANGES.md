Changelog
==========

# 0.2.0 (2015-04-16)

Core Stability and Driver Updates

## Core

- Support for system proxy environment
- New command to regenerate TLS certificates
  - Note: this will restart the Docker engine to apply
- Updates to driver operations (create, start, stop, etc) for better reliability
- New internal `libmachine` package for internal api (not ready for public usage)
- Updated Driver Interface
  - [Driver Spec](https://github.com/docker/machine/blob/master/docs/DRIVER_SPEC.md)
  - Removed host provisioning from Drivers to enable a more consistent install
  - Removed SSH commands from each Driver for more consistent operations
- Swarm: machine now uses Swarm default binpacking strategy

## Driver Updates

- All drivers updated to new Driver interface
- Amazon EC2
  - Better checking for subnets on creation
  - Support for using Private IPs in VPC
  - Fixed bug with duplicate security group authorization with Swarm
  - Support for IAM instance profile
  - Fixed bug where IP was not properly detected upon stop
- DigitalOcean
  - IPv6 support
  - Backup option
  - Private Networking
- Openstack / Rackspace
  - Gophercloud updated to latest version
  - New insecure flag to disable TLS (use with caution)
- Google
  - Google source image updated
  - Ability to specify auth token via file
- VMware Fusion
  - Paravirtualized driver for disk (pvscsi)
  - Enhanced paravirtualized NIC (vmxnet3)
  - Power option updates
  - SSH keys persistent across reboots
  - Stop now gracefully stops VM
  - vCPUs now match host CPUs
- Softlayer
  - Fixed provision bug where `curl` was not present
- VirtualBox
  - Correct power operations with Saved VM state
  - Fixed bug where image option was ignored

## CLI

- Auto-regeneration of TLS certificates when TLS error is detected
  - Note: this will restart the Docker engine to apply
- Minor UI updates including improved sorting and updated command docs
- Bug with `config` and `env` with spaces fixed
  - Note: you now must use `eval $(docker-machine env machine)` to load environment settings
- Updates to better support `fish` shell
- Use `--tlsverify` for both `config` and `env` commands
- Commands now use eval for better interoperability with shell

## Testing
- New integration test framework (bats)


# 0.1.0 (2015-02-26)

Initial beta release.

- Provision Docker Engines using multiple drivers
- Provide light management for the machines
  - Create, Start, Stop, Restart, Kill, Remove, SSH
- Configure the Docker Engine for secure communication (TLS)
- Easily switch target machine for fast configuration of Docker Engine client
- Provision Swarm clusters (experimental)

## Included drivers

- Amazon EC2
- Digital Ocean
- Google
- Microsoft Azure
- Microsoft Hyper-V
- Openstack
- Rackspace
- VirtualBox
- VMware Fusion
- VMware vCloud Air
- VMware vSphere

Changelog
==========

# 0.3.0 (2015-06-18)

## Features
- Engine option configuration (ability to configure all engine options)
- Swarm option configuration (ability to configure all swarm options)
- New Provisioning system to allow for greater flexibility and stability for installing and configuring Docker
- New Provisioners
  - Rancher OS
  - RedHat Enterprise Linux 7.0+ (experimental)
  - Fedora 21+ (experimental)
  - Debian 8+ (experimental)
- PowerShell support (configure Windows Docker CLI)
- Command Prompt (cmd.exe) support (configure Windows Docker CLI)
- Filter command help by driver
- Ability to import Boot2Docker instances
- Boot2Docker CLI migration guide (experimental)
- Format option for `inspect` command
- New logging output format to improve readability and display across platforms 
- Updated "active" machine concept - now is implicit according to `DOCKER_HOST` environment variable.  Note: this removes the implicit "active" machine and can no longer be specified with the `active` command.  You change the "active" host by using the `env` command instead.
- Specify Swarm version (`--swarm-image` flag)

## Drivers
- New: Exoscale Driver
- New: Generic Driver (provision any host with supported base OS and SSH)
- Amazon EC2
  - SSH user is configurable
  - Support for Spot instances
  - Add option to use private address only
  - Base AMI updated to 20150417
- Google
  - Support custom disk types
  - Updated base image to v20150316
- Openstack
  - Support for Keystone v3 domains
- Rackspace
  - Misc fixes including environment variable for Flavor Id and stability
- Softlayer
  - Enable local disk as provisioning option
  - Fixes for SSH access errors
  - Fixed bug where public IP would always be returned when requesting private
  - Add support for specifying public and private VLAN IDs
- VirtualBox
  - Use Intel network interface driver (adds great stability)
  - Stability fixes for NAT access
  - Use DNS pass through
  - Default CPU to single core for improved performance
  - Enable shared folder support for Windows hosts
- VMware Fusion
  - Boot2Docker ISO updated
  - Shared folder support

## Fixes
- Provisioning improvements to ensure Docker is available
- SSH improvements for provisioning stability
- Fixed SSH key generation bug on Windows
- Help formatting for improved readability

## Breaking Changes
- "Short-Form" name reference no longer supported Instead of "docker-machine " implying the active host you must now use docker-machine
- VMware shared folders require Boot2Docker 1.7

## Special Thanks                                                                        
We would like to thank all contributors.  Machine would not be where it is
without you.  We would also like to give special thanks to the following
contributors for outstanding contributions to the project:

- @frapposelli for VMware updates and fixes
- @hairyhenderson for several improvements to Softlayer driver, inspect formatting and lots of fixes
- @ibuildthecloud for rancher os provisioning
- @sthulb for portable SSH library                                              
- @vincentbernat for exoscale                                                   
- @zchee for Amazon updates and great doc updates

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
- SoftLayer
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

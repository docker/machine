<!--[metadata]>
+++
title = "Oracle VirtualBox"
description = "Oracle VirtualBox driver for machine"
keywords = ["machine, Oracle VirtualBox, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Oracle VirtualBox
Create machines locally using [VirtualBox](https://www.virtualbox.org/).
This driver requires VirtualBox to be installed on your host.

    $ docker-machine create --driver=virtualbox vbox-test
    
You can create an entirely new machine or you can convert a Boot2Docker VM into
a machine by importing the VM. To convert a Boot2Docker VM, you'd use the following
command:

    $ docker-machine create -d virtualbox --virtualbox-import-boot2docker-vm boot2docker-vm b2d


Options:

 - `--virtualbox-memory`: Size of memory for the host in MB.
 - `--virtualbox-cpu-count`: Number of CPUs to use to create the VM. Defaults to single CPU.
 - `--virtualbox-disk-size`: Size of disk for the host in MB.
 - `--virtualbox-boot2docker-url`: The URL of the boot2docker image. Defaults to the latest available version.
 - `--virtualbox-import-boot2docker-vm`: The name of a Boot2Docker VM to import.
 - `--virtualbox-hostonly-cidr`: The CIDR of the host only adapter.

The `--virtualbox-boot2docker-url` flag takes a few different forms. By
default, if no value is specified for this flag, Machine will check locally for
a boot2docker ISO. If one is found, that will be used as the ISO for the
created machine. If one is not found, the latest ISO release available on
[boot2docker/boot2docker](https://github.com/boot2docker/boot2docker) will be
downloaded and stored locally for future use. Note that this means you must run
`docker-machine upgrade` deliberately on a machine if you wish to update the "cached"
boot2docker ISO.

This is the default behavior (when `--virtualbox-boot2docker-url=""`), but the
option also supports specifying ISOs by the `http://` and `file://` protocols.
`file://` will look at the path specified locally to locate the ISO: for
instance, you could specify `--virtualbox-boot2docker-url
file://$HOME/Downloads/rc.iso` to test out a release candidate ISO that you have
downloaded already. You could also just get an ISO straight from the Internet
using the `http://` form.

To customize the host only adapter, you can use the `--virtualbox-hostonly-cidr`
flag.  This will specify the host IP and Machine will calculate the VirtualBox
DHCP server address (a random IP on the subnet between `.1` and `.25`) so 
it does not clash with the specified host IP.
Machine will also specify the DHCP lower bound to `.100` and the upper bound
to `.254`.  For example, a specified CIDR of `192.168.24.1/24` would have a
DHCP server between `192.168.24.2-25`, a lower bound of `192.168.24.100` and 
upper bound of `192.168.24.254`.

Environment variables and default values:

| CLI option                           | Environment variable         | Default                  |
|--------------------------------------|------------------------------|--------------------------|
| `--virtualbox-memory`                | `VIRTUALBOX_MEMORY_SIZE`     | `1024`                   |
| `--virtualbox-cpu-count`             | `VIRTUALBOX_CPU_COUNT`       | `1`                      |
| `--virtualbox-disk-size`             | `VIRTUALBOX_DISK_SIZE`       | `20000`                  |
| `--virtualbox-boot2docker-url`       | `VIRTUALBOX_BOOT2DOCKER_URL` | *Latest boot2docker url* |
| `--virtualbox-import-boot2docker-vm` | -                            | `boot2docker-vm`         |
| `--virtualbox-hostonly-cidr`         | `VIRTUALBOX_HOSTONLY_CIDR`   | `192.168.99.1/24`        |

<!--[metadata]>
+++
title = "VMware Fusion"
description = "VMware Fusion driver for machine"
keywords = ["machine, VMware Fusion, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# VMware Fusion
Creates machines locally on [VMware Fusion](http://www.vmware.com/products/fusion). Requires VMware Fusion to be installed.

Options:

 - `--vmwarefusion-boot2docker-url`: URL for boot2docker image.
 - `--vmwarefusion-cpu-count`: Number of CPUs for the machine (-1 to use the number of CPUs available)
 - `--vmwarefusion-disk-size`: Size of disk for host VM (in MB).
 - `--vmwarefusion-memory-size`: Size of memory for host VM (in MB).

The VMware Fusion driver uses the latest boot2docker image.
See [frapposelli/boot2docker](https://github.com/frapposelli/boot2docker/tree/vmware-64bit)

Environment variables and default values:

| CLI option                       | Environment variable     | Default                  |
|----------------------------------|--------------------------|--------------------------|
| `--vmwarefusion-boot2docker-url` | `FUSION_BOOT2DOCKER_URL` | *Latest boot2docker url* |
| `--vmwarefusion-cpu-count`       | `FUSION_CPU_COUNT`       | `1`                      |
| `--vmwarefusion-disk-size`       | `FUSION_DISK_SIZE`       | `20000`                  |
| `--vmwarefusion-memory-size`     | `FUSION_MEMORY_SIZE`     | `1024`                   |

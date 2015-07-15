<!--[metadata]>
+++
title = "VMware vSphere"
description = "VMware vSphere driver for machine"
keywords = ["machine, VMware vSphere, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# VMware vSphere
Creates machines on a [VMware vSphere](http://www.vmware.com/products/vsphere) Virtual Infrastructure. The machine must have a working vSphere ESXi installation. You can use a paid license or free 60 day trial license. Your installation may also include an optional VCenter server. The vSphere driver depends on [`govc`](https://github.com/vmware/govmomi/tree/master/govc) (must be in path) and has been tested with [vmware/govmomi@`c848630`](https://github.com/vmware/govmomi/commit/c8486300bfe19427e4f3226e3b3eac067717ef17).

Options:

 - `--vmwarevsphere-cpu-count`: CPU number for Docker VM.
 - `--vmwarevsphere-memory-size`: Size of memory for Docker VM (in MB).
 - `--vmwarevsphere-boot2docker-url`: URL for boot2docker image.
 - `--vmwarevsphere-vcenter`: IP/hostname for vCenter (or ESXi if connecting directly to a single host).
 - `--vmwarevsphere-disk-size`: Size of disk for Docker VM (in MB).
 - `--vmwarevsphere-username`: **required** vSphere Username.
 - `--vmwarevsphere-password`: **required** vSphere Password.
 - `--vmwarevsphere-network`: Network where the Docker VM will be attached.
 - `--vmwarevsphere-datastore`: Datastore for Docker VM.
 - `--vmwarevsphere-datacenter`: Datacenter for Docker VM (must be set to `ha-datacenter` when connecting to a single host).
 - `--vmwarevsphere-pool`: Resource pool for Docker VM.
 - `--vmwarevsphere-compute-ip`: Compute host IP where the Docker VM will be instantiated.

The VMware vSphere driver uses the latest boot2docker image.

Environment variables and default values:

| CLI option                        | Environment variable      | Default                  |
|-----------------------------------|---------------------------|--------------------------|
| `--vmwarevsphere-cpu-count`       | `VSPHERE_CPU_COUNT`       | `2`                      |
| `--vmwarevsphere-memory-size`     | `VSPHERE_MEMORY_SIZE`     | `2048`                   |
| `--vmwarevsphere-disk-size`       | `VSPHERE_DISK_SIZE`       | `20000`                  |
| `--vmwarevsphere-boot2docker-url` | `VSPHERE_BOOT2DOCKER_URL` | *Latest boot2docker url* |
| `--vmwarevsphere-vcenter`         | `VSPHERE_VCENTER`         | -                        |
| **`--vmwarevsphere-username`**    | `VSPHERE_USERNAME`        | -                        |
| **`--vmwarevsphere-password`**    | `VSPHERE_PASSWORD`        | -                        |
| `--vmwarevsphere-network`         | `VSPHERE_NETWORK`         | -                        |
| `--vmwarevsphere-datastore`       | `VSPHERE_DATASTORE`       | -                        |
| `--vmwarevsphere-datacenter`      | `VSPHERE_DATACENTER`      | -                        |
| `--vmwarevsphere-pool`            | `VSPHERE_POOL`            | -                        |
| `--vmwarevsphere-compute-ip`      | `VSPHERE_COMPUTE_IP`      | -                        |
<!--[metadata]>
+++
title = "Microsoft Hyper-V"
description = "Microsoft Hyper-V driver for machine"
keywords = ["machine, Microsoft Hyper-V, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Microsoft Hyper-V

Creates a Boot2Docker virtual machine locally on your Windows machine
using Hyper-V. [See here](http://windows.microsoft.com/en-us/windows-8/hyper-v-run-virtual-machines)
for instructions to enable Hyper-V. You will need to use an
Administrator level account to create and manage Hyper-V machines.

> **Note**: You will need an existing virtual switch to use the
> driver. Hyper-V can share an external network interface (aka
> bridging), see [this blog](http://blogs.technet.com/b/canitpro/archive/2014/03/11/step-by-step-enabling-hyper-v-for-use-on-windows-8-1.aspx).
> If you would like to use NAT, create an internal network, and use
> [Internet Connection
> Sharing](http://www.packet6.com/allowing-windows-8-1-hyper-v-vm-to-work-with-wifi/).

Options:

-   `--hyperv-boot2docker-url`: The URL of the boot2docker ISO.
-   `--hyperv-virtual-switch`: Name of the virtual switch to use.
-   `--hyperv-disk-size`: Size of disk for the host in MB.
-   `--hyperv-memory`: Size of memory for the host in MB.

Environment variables and default values:

| CLI option                      | Environment variable | Default                  |
| ------------------------------- | -------------------- | ------------------------ |
| `--hyperv-boot2docker-url`      | -                    | _Latest boot2docker url_ |
| `--hyperv-virtual-switch`       | -                    | _first found_            |
| `--hyperv-disk-size`            | -                    | `20000`                  |
| `--hyperv-memory`               | -                    | `1024`                   |

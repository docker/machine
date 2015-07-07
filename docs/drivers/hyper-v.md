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

 - `--hyper-v-boot2docker-url`: The URL of the boot2docker ISO. Defaults to the latest available version.
 - `--hyper-v-boot2docker-location`: Location of a local boot2docker iso to use. Overrides the URL option below.
 - `--hyper-v-virtual-switch`: Name of the virtual switch to use. Defaults to first found.
 - `--hyper-v-disk-size`: Size of disk for the host in MB.
 - `--hyper-v-memory`: Size of memory for the host in MB. By default, the machine is setup to use dynamic memory.

Environment variables and default values:

| CLI option                       | Environment variable | Default                  |
|----------------------------------|----------------------| -------------------------|
| `--hyper-v-boot2docker-url`      | -                    | *Latest boot2docker url* |
| `--hyper-v-boot2docker-location` | -                    | -                        |
| `--hyper-v-virtual-switch`       | -                    | *first found*            |
| `--hyper-v-disk-size`            | -                    | `20000`                  |
| `--hyper-v-memory`               | -                    | `1024`                   |

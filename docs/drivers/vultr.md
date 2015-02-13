<!--[metadata]>
+++
title = "Vultr"
description = "Vultr driver for machine"
keywords = ["machine, Vultr, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Digital Ocean
Create machines on [Vultr](https://www.vultr.com/).

Get your API key from the [Vultr control panel](https://my.vultr.com/settings/) and pass
that to `docker-machine create` with the `--vultr-api-key` option.

    $ docker-machine create --driver vultr --vultr-api-key=aa9399a2175a93b17b1c86c807e08d3fc4b test-vps

Options:

 - `--vultr-api-key`: **required** Your Vultr API key.
 - `--vultr-os-id`: Operating system ID (OSID) to use. See [API OS endpoint](https://www.vultr.com/api/#os_os_list).
 - `--vultr-region-id`: ID of the region to create the VPS in. See [API Region endpoint](https://www.vultr.com/api/#regions_region_list).
 - `--vultr-plan-id`: Plan ID (VPSPLANID).
 - `--vultr-ipv6`: Enable IPv6 support for the VPS.
 - `--vultr-private-networking`: Enable private networking support for the VPS.
 - `--vultr-backups`: Enable automatic backups for the VPS.

The Vultr driver uses OS ID `160` (Ubuntu 14.04 x64) by default.     
Since the deployment of an Ubuntu machine on Vultr can take several minutes, you can alternatively do a PXE-based deployment of [RancherOS](http://rancher.com/rancher-os/) by supplying the following flag:

    --vultr-os-id=159

 Environment variables and default values:

| CLI option                      | Environment variable         | Default            |
|---------------------------------|------------------------------|--------------------|
| **`--vultr-api-key`**           | `VULTR_API_KEY`              | -                  |
| `--vultr-os-id`                 | `VULTR_OS`                   | *Ubuntu 14.04 x64* |
| `--vultr-region-id`             | `VULTR_REGION`               | *New Jersey*       |
| `--vultr-plan-id`               | `VULTR_PLAN`                 | *768 MB RAM*       |
| `--vultr-ipv6`                  | `VULTR_IPV6`                 | `false`            |
| `--vultr-private-networking`    | `VULTR_PRIVATE_NETWORKING`   | `false`            |
| `--vultr-backups`               | `VULTR_BACKUPS`              | `false`            |

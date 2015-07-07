<!--[metadata]>
+++
title = "exoscale"
description = "exoscale driver for machine"
keywords = ["machine, exoscale, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Exoscale
Create machines on [exoscale](https://www.exoscale.ch/).

Get your API key and API secret key from [API details](https://portal.exoscale.ch/account/api) and pass them to `machine create` with the `--exoscale-api-key` and `--exoscale-api-secret-key` options.

Options:

 - `--exoscale-url`: Your API endpoint.
 - `--exoscale-api-key`: **required** Your API key.
 - `--exoscale-api-secret-key`: **required** Your API secret key.
 - `--exoscale-instance-profile`: Instance profile.
 - `--exoscale-disk-size`: Disk size for the host in GB.
 - `--exoscale-image`: exoscale disk size. (10, 50, 100, 200, 400)
 - `--exoscale-security-group`: Security group. It will be created if it doesn't exist.
 - `--exoscale-availability-zone`: exoscale availability zone.

If a custom security group is provided, you need to ensure that you allow TCP ports 22 and 2376 in an ingress rule. Moreover, if you want to use Swarm, also add TCP port 3376.

Environment variables and default values:

| CLI option                      | Environment variable         | Default                           |
|---------------------------------|------------------------------|-----------------------------------|
| `--exoscale-url`                | `EXOSCALE_ENDPOINT`          | `https://api.exoscale.ch/compute` |
| **`--exoscale-api-key`**        | `EXOSCALE_API_KEY`           | -                                 |
| **`--exoscale-api-secret-key`** | `EXOSCALE_API_SECRET`        | -                                 |
| `--exoscale-instance-profile`   | `EXOSCALE_INSTANCE_PROFILE`  | `small`                           |
| `--exoscale-disk-size`          | `EXOSCALE_DISK_SIZE`         | `50`                              |
| `--exoscale-image`              | `EXSOCALE_IMAGE`             | `ubuntu-14.04`                    |
| `--exoscale-security-group`     | `EXOSCALE_SECURITY_GROUP`    | `docker-machine`                  |
| `--exoscale-availability-zone`  | `EXOSCALE_AVAILABILITY_ZONE` | `ch-gva-2`                        |
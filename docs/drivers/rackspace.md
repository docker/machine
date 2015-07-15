<!--[metadata]>
+++
title = "Rackspace"
description = "Rackspace driver for machine"
keywords = ["machine, Rackspace, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Rackspace
Create machines on [Rackspace cloud](http://www.rackspace.com/cloud)

Options:

 - `--rackspace-username`: **required** Rackspace account username.
 - `--rackspace-api-key`: **required** Rackspace API key.
 - `--rackspace-region`: **required** Rackspace region name.
 - `--rackspace-endpoint-type`: Rackspace endpoint type (`adminURL`, `internalURL` or the default `publicURL`).
 - `--rackspace-image-id`: Rackspace image ID. Default: Ubuntu 14.10 (Utopic Unicorn) (PVHVM).
 - `--rackspace-flavor-id`: Rackspace flavor ID. Default: General Purpose 1GB.
 - `--rackspace-ssh-user`: SSH user for the newly booted machine.
 - `--rackspace-ssh-port`: SSH port for the newly booted machine.
 - `--rackspace-docker-install`: Set if Docker has to be installed on the machine.

The Rackspace driver will use `598a4282-f14b-4e50-af4c-b3e52749d9f9` (Ubuntu 14.04 LTS) by default.

Environment variables and default values:

| CLI option                   | Environment variable | Default                                |
|------------------------------|----------------------|----------------------------------------|
| **`--rackspace-username`**   | `OS_USERNAME`        | -                                      |
| **`--rackspace-api-key`**    | `OS_API_KEY`         | -                                      |
| **`--rackspace-region`**     | `OS_REGION_NAME`     | -                                      |
| `--rackspace-endpoint-type`  | `OS_ENDPOINT_TYPE`   | `publicURL`                            |
| `--rackspace-image-id`       | -                    | `598a4282-f14b-4e50-af4c-b3e52749d9f9` |
| `--rackspace-flavor-id`      | `OS_FLAVOR_ID`       | `general1-1`                           |
| `--rackspace-ssh-user`       | -                    | `root`                                 |
| `--rackspace-ssh-port`       | -                    | `22`                                   |
| `--rackspace-docker-install` | -                    | `true`                                 |

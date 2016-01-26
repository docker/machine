<!--[metadata]>
+++
title = "Generic"
description = "Generic driver for machine"
keywords = ["machine, Generic, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Generic

Create machines using an existing VM/Host with SSH.

This is useful if you are using a provider that Machine does not support
directly or if you would like to import an existing host to allow Docker
Machine to manage.

The driver will perform a list of tasks on create:

-   If docker is not running on the host, it will be installed automatically.
-   It will generate certificates to secure the docker daemon
-   The docker daemon will be restarted, thus all running containers will be stopped.


    $ docker-machine create --driver generic --generic-ip-address=203.0.113.81 vm

Options:

-   `--generic-ip-address`: **required** IP Address of host.
-   `--generic-ssh-key`: Path to the SSH user private key.
-   `--generic-ssh-user`: SSH username used to connect.
-   `--generic-ssh-port`: Port to use for SSH.

> **Note**: You must use a base operating system supported by Machine.

Environment variables and default values:

| CLI option                 | Environment variable | Default                   |
| -------------------------- | -------------------- | ------------------------- |
| **`--generic-ip-address`** | `GENERIC_IP_ADDRESS` | -                         |
| `--generic-ssh-key`        | `GENERIC_SSH_KEY`    | _(defers to `ssh-agent`)_ |
| `--generic-ssh-user`       | `GENERIC_SSH_USER`   | `root`                    |
| `--generic-ssh-port`       | `GENERIC_SSH_PORT`   | `22`                      |

##### Interaction with SSH Agents

When an SSH identity is not provided (with the `--generic-ssh-key` flag),
the SSH agent (if running) will be consulted. This makes it possible to
easily use password-protected SSH keys.

Note that this usage is _only_ supported if you're using the external SSH client,
which is the default behaviour when the `ssh` binary is available. If you're
using the native client (with `--native-ssh`), using the SSH agent is not yet
supported.

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

Options:

-   `--generic-ip-address`: **required** IP Address of host.
-   `--generic-ssh-key`: Path to the SSH user private key.
-   `--generic-ssh-user`: SSH username used to connect.
-   `--generic-ssh-port`: Port to use for SSH.

> **Note**: You must use a base operating system supported by Machine.

Environment variables and default values:

| CLI option                 | Environment variable | Default             |
| -------------------------- | -------------------- | ------------------- |
| **`--generic-ip-address`** | `GENERIC_IP_ADDRESS` | -                   |
| `--generic-ssh-key`        | `GENERIC_SSH_KEY`    | `$HOME/.ssh/id_rsa` |
| `--generic-ssh-user`       | `GENERIC_SSH_USER`   | `root`              |
| `--generic-ssh-port`       | `GENERIC_SSH_PORT`   | `22`                |

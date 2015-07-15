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

Options:

 - `--generic-ip-address`: **required** IP Address of host.
 - `--generic-ssh-user`: SSH username used to connect.
 - `--generic-ssh-key`: Path to the SSH user private key.
 - `--generic-ssh-port`: Port to use for SSH.

> **Note**: You must use a base operating system supported by Machine.

Environment variables and default values:

| CLI option                 | Environment variable | Default             |
|----------------------------|----------------------|---------------------|
| **`--generic-ip-address`** | -                    | -                   |
| `--generic-ssh-user`       | -                    | `root`              |
| `--generic-ssh-key`        | -                    | `$HOME/.ssh/id_rsa` |
| `--generic-ssh-port`       | -                    | `22`                |

<!--[metadata]>
+++
title = "Parallels Desktop for Mac"
description = "Parallels Desktop driver for machine"
keywords = ["machine, Parallels Desktop, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

#### Parallels
Creates machines locally on [Parallels Desktop for Mac](http://www.parallels.com/products/desktop/).
Requires _Parallels Desktop for Mac_ version 11 or higher to be installed.

    $ docker-machine create --driver=parallels prl-test

Options:

 - `--parallels-boot2docker-url`: The URL of the boot2docker image.
 - `--parallels-disk-size`: Size of disk for the host VM (in MB).
 - `--parallels-memory`: Size of memory for the host VM (in MB).
 - `--parallels-cpu-count`: Number of CPUs to use to create the VM (-1 to use the number of CPUs available).

The `--parallels-boot2docker-url` flag takes a few different forms. By
default, if no value is specified for this flag, Machine will check locally for
a boot2docker ISO. If one is found, that will be used as the ISO for the
created machine. If one is not found, the latest ISO release available on
[boot2docker/boot2docker](https://github.com/boot2docker/boot2docker) will be
downloaded and stored locally for future use. Note that this means you must run
`docker-machine upgrade` deliberately on a machine if you wish to update the "cached"
boot2docker ISO.

This is the default behavior (when `--parallels-boot2docker-url=""`), but the
option also supports specifying ISOs by the `http://` and `file://` protocols.

Environment variables and default values:

| CLI option                    | Environment variable        | Default                  |
|-------------------------------|-----------------------------|--------------------------|
| `--parallels-boot2docker-url` | `PARALLELS_BOOT2DOCKER_URL` | *Latest boot2docker url* |
| `--parallels-cpu-count`       | `PARALLELS_CPU_COUNT`       | `1`                      |
| `--parallels-disk-size`       | `PARALLELS_DISK_SIZE`       | `20000`                  |
| `--parallels-memory`          | `PARALLELS_MEMORY_SIZE`     | `1024`                   |

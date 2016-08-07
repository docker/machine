<!--[metadata]>
+++
title = "GoDaddy Cloud Servers"
description = "GoDaddy Cloud Servers driver for machine"
keywords = ["machine, GoDaddy Cloud Servers, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# GoDaddy Cloud Servers

Create machines on [GoDaddy Cloud Servers](https://www.godaddy.com/pro/cloud-servers).

To create machines on [GoDaddy Cloud Servers](https://www.godaddy.com/pro/cloud-servers)
you will need an API key associated with your GoDaddy account. API keys can be obtained
on the [GoDaddy Developer Portal](https://developer.godaddy.com/keys/).

With an API key in hand, you can create a new server with:

```bash
$ docker-machine create --driver godaddy --godaddy-api-key=${APIKEY} myhost
```

Alternatively, you can use environment variables:

```bash
$ export GODADDY_API_KEY=${APIKEY}
$ docker-machine create -d godaddy myhost
```

Options:

-   `--godaddy-api-key`: Your GoDaddy API key.
-   `--godaddy-base-api-url`: The GoDaddy API endpoint.
-   `--godaddy-boot-timeout`: Amount of time (in seconds) to wait for the initial boot.
-   `--godaddy-image`: The image to use for the new server.
-   `--godaddy-spec`: The server spec to use for the new server.
-   `--godaddy-ssh-key`: Private keyfile to use for SSH (absolute path).
-   `--godaddy-ssh-key-id`: Id of the existing GoDaddy SSH Key to associate with this new server.
-   `--godaddy-data-center`: The GoDaddy data center to launch the server in.
-   `--godaddy-zone`: The data center zone to launch the server in.
-   `--godaddy-ssh-user`: Name of the user to be used for SSH.



Environment variables and default values:

| CLI option                      | Environment variable         | Default                           |
| ------------------------------- | ---------------------------- | --------------------------------- |
| `--godaddy-api-key`             | `GODADDY_API_KEY`            | -                                 |
| `--godaddy-base-api-url`        | `GODADDY_API_URL`            | `https://api.godaddy.com`         |
| `--godaddy-boot-timeout`        | `GODADDY_BOOT_TIMEOUT`       | `120`                             |
| `--godaddy-image`               | `GODADDY_IMAGE`              | `ubuntu-1604`                     |
| `--godaddy-spec`                | `GODADDY_SPEC`               | `tiny`                            |
| `--godaddy-ssh-key`             | `GODADDY_SSH_KEY`            | -                                 |
| `--godaddy-ssh-key-id`          | `GODADDY_SSH_KEY_ID`         | -                                 |
| `--godaddy-data-center`         | `GODADDY_DATA_CENTER`        | `phx`                             |
| `--godaddy-zone`                | `GODADDY_ZONE`               | `phx-1`                           |
| `--godaddy-ssh-user`            | `GODADDY_SSH_USER`           | `machine`                         |

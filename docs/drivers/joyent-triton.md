<!--[metadata]>
+++
title = "Joyent Triton"
description = "Joyent Triton driver for machine"
keywords = ["machine, triton, driver, joyent"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Joyent Triton

Create machines on [Joyent's Triton](http://joyent.com/) Elastic Container Infrastructure.

Before using the Docker service on Triton, you need to have completed the signup
process and generated a set of SSH keys (make sure you know the account name and
location of your SSH key). The signup process will also provide you with the
necessary CloudAPI url you will use to create the docker-machine entry. Please visit
[getting started](https://www.joyent.com/developers/getting-started) for more
information.

To create the docker-machine configuration:

    1. Ensure you have *openssl* and *ssh-keygen* installed
    2. Run docker-mache create:

        $ docker-machine create --driver triton --triton-url=$CLOUDAPI_URL --triton-account=$ACCOUNT --triton-key=$PATH_TO_PRIVATE_SSH_KEY  $NAME

During the creation process, your private key will be used to generate a
certificate that the Docker client can then use to communicate securely with the
Triton docker service.

To use COAL (Joyent's Triton on a laptop) please visit [the getting
started with Cloud on a Laptop section](https://github.com/joyent/sdc#getting-started),
else if you are running your own Triton SmartDataCenter [here](https://github.com/joyent/sdc-docker)
has everything you need.

Note that some docker-machine commands are not supported in Triton (e.g. ssh,
start, stop, upgrade) as Triton provides a dynamic and flexible docker cloud
environment, and so the provisioning of individual machines isn't necessary.

Triton driver options:

 - `--triton-url`             : Required - the Triton Cloudapi URL (e.g. https://us-east-3b.api.joyent.com).
 - `--triton-account`         : Required - the Triton acount name (e.g. jill).
 - `--triton-key`             : Required - the path to your private key file (e.g. /Users/jill/.ssh/id_rsa).
 - `--triton-datacenter`      : Optional - short datacenter name to be used instead of --triton-url (e.g. us-east-3b).
 - `--triton-skip-tls-verify` : Optional - to skip the TLS server verification (insecure)

Environment variables and default values:

| CLI option                    | Environment variable   | Default                          |
|-------------------------------|------------------------|----------------------------------|
| `--triton-url`                | `SDC_URL`              | https://us-east-1.api.joyent.com |
| `--triton-account`            | `SDC_ACCOUNT`          | -                                |
| `--triton-key`                | `SDC_KEY`              | ~/.ssh/id_rsa                    |
| `--triton-datacenter`         | `SDC_DC`               | -                                |
| `--triton-skip-tls-verify`    | `SDC_SKIP_TLS_VERIFY`  | false                            |

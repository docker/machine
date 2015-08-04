<!--[metadata]>
+++
title = "OpenStack"
description = "OpenStack driver for machine"
keywords = ["machine, OpenStack, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# OpenStack
Create machines on [OpenStack](http://www.openstack.org/software/)

Mandatory:

 - `--openstack-auth-url`: Keystone service base URL.
 - `--openstack-flavor-id` or `--openstack-flavor-name`: Identify the flavor that will be used for the machine.
 - `--openstack-image-id` or `--openstack-image-name`: Identify the image that will be used for the machine.

Options:

 - `--openstack-insecure`: Explicitly allow openstack driver to perform "insecure" SSL (https) requests. The server's certificate will not be verified against any certificate authorities. This option should be used with caution.
 - `--openstack-domain-name` or `--openstack-domain-id`: Domain to use for authentication (Keystone v3 only)
 - `--openstack-username`: User identifier to authenticate with.
 - `--openstack-password`: User password. It can be omitted if the standard environment variable `OS_PASSWORD` is set.
 - `--openstack-tenant-name` or `--openstack-tenant-id`: Identify the tenant in which the machine will be created.
 - `--openstack-region`: The region to work on. Can be omitted if there is only one region on the OpenStack.
 - `--openstack-availability-zone`: The availability zone in which to launch the server.
 - `--openstack-endpoint-type`: Endpoint type can be `internalURL`, `adminURL` on `publicURL`. If is a helper for the driver
   to choose the right URL in the OpenStack service catalog. If not provided the default id `publicURL`
 - `--openstack-net-name` or `--openstack-net-id`: Identify the private network the machine will be connected on. If your OpenStack project project contains only one private network it will be use automatically.
 - `--openstack-sec-groups`: If security groups are available on your OpenStack you can specify a comma separated list
   to use for the machine (e.g. `secgrp001,secgrp002`).
 - `--openstack-floatingip-pool`: The IP pool that will be used to get a public IP can assign it to the machine. If there is an
   IP address already allocated but not assigned to any machine, this IP will be chosen and assigned to the machine. If
   there is no IP address already allocated a new IP will be allocated and assigned to the machine.
 - `--openstack-ssh-user`: The username to use for SSH into the machine. If not provided `root` will be used.
 - `--openstack-ssh-port`: Customize the SSH port if the SSH server on the machine does not listen on the default port.

Environment variables and default values:

| CLI option                       | Environment variable   | Default |
|----------------------------------|------------------------|---------|
| `--openstack-auth-url`           | `OS_AUTH_URL`          | -       |
| `--openstack-flavor-name`        | -                      | -       |
| `--openstack-flavor-id`          | -                      | -       |
| `--openstack-image-name`         | -                      | -       |
| `--openstack-image-id`           | -                      | -       |
| `--openstack-insecure`           | -                      | -       |
| `--openstack-domain-name`        | `OS_DOMAIN_NAME`       | -       |
| `--openstack-domain-id`          | `OS_DOMAIN_ID`         | -       |
| `--openstack-username`           | `OS_USERNAME`          | -       |
| `--openstack-password`           | `OS_PASSWORD`          | -       |
| `--openstack-tenant-name`        | `OS_TENANT_NAME`       | -       |
| `--openstack-tenant-id`          | `OS_TENANT_ID`         | -       |
| `--openstack-region`             | `OS_REGION_NAME`       | -       |
| `--openstack-availability-zone`  | `OS_AVAILABILITY_ZONE` | -       |
| `--openstack-endpoint-type`      | `OS_ENDPOINT_TYPE`     | -       |
| `--openstack-net-name`           | -                      | -       |
| `--openstack-net-id`             | -                      | -       |
| `--openstack-sec-groups`         | -                      | -       |
| `--openstack-floatingip-pool`    | -                      | -       |
| `--openstack-ssh-user`           | -                      | `root`  |
| `--openstack-ssh-port`           | -                      | `22`    |

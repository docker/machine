<!--[metadata]>
+++
title = "Google Compute Engine"
description = "Google Compute Engine driver for machine"
keywords = ["machine, Google Compute Engine, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Google Compute Engine

Create machines on [Google Compute Engine](https://cloud.google.com/compute/). You will need a Google account and project name. See https://cloud.google.com/compute/docs/projects for details on projects.

The Google driver uses oAuth. When creating the machine, you will have your browser opened to authorize. Once authorized, paste the code given in the prompt to launch the instance.

Options:

 - `--google-zone`: The zone to launch the instance.
 - `--google-machine-type`: The type of instance.
 - `--google-username`: The username to use for the instance.
 - `--google-project`: **required** The name of your project to use when launching the instance.
 - `--google-auth-token`: Your oAuth token for the Google Cloud API.
 - `--google-scopes`: The scopes for OAuth 2.0 to Access Google APIs. See [Google Compute Engine Doc](https://cloud.google.com/storage/docs/authentication).
 - `--google-disk-size`: The disk size of instance.
 - `--google-disk-type`: The disk type of instance.

The GCE driver will use the `ubuntu-1404-trusty-v20150316` instance type unless otherwise specified.

Environment variables and default values:

| CLI option                | Environment variable  | Default                              |
|---------------------------|-----------------------|--------------------------------------|
| `--google-zone`           | `GOOGLE_ZONE`         | `us-central1-a`                      |
| `--google-machine-type`   | `GOOGLE_MACHINE_TYPE` | `f1-micro`                           |
| `--google-username`       | `GOOGLE_USERNAME`     | `docker-user`                        |
| **`--google-project`**    | `GOOGLE_PROJECT`      | -                                    |
| `--google-auth-token`     | `GOOGLE_AUTH_TOKEN`   | -                                    |
| `--google-scopes`         | `GOOGLE_SCOPES`       | `devstorage.read_only,logging.write` |
| `--google-disk-size`      | `GOOGLE_DISK_SIZE`    | `10`                                 |
| `--google-disk-type`      | `GOOGLE_DISK_TYPE`    | `pd-standard`                        |

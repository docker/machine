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

Create machines on [Google Compute Engine](https://cloud.google.com/compute/).
You will need a Google account and a project id.
See https://cloud.google.com/compute/docs/projects for details on projects.

### Credentials

The Google driver uses [Application Default Credentials](https://developers.google.com/identity/protocols/application-default-credentials)
to get authorization credentials for use in calling Google APIs.

So if `docker-machine` is used from a GCE host, authentication will happen automatically
via the built-in service account.
Otherwise, [install gcloud](https://cloud.google.com/sdk/) and get
through the oauth2 process with `gcloud auth login`.

### Example

To create a machine instance, specify `--driver google`, the project id and the machine name.

```
$ gcloud auth login
$ docker-machine create --driver google --google-project PROJECT_ID vm01
$ docker-machine create --driver google \
  --google-project PROJECT_ID \
  --google-zone us-central1-a \
  --google-machine-type f1-micro \
  vm02
```

### Options

- `--google-project`: **required** The id of your project to use when launching the instance.
 - `--google-zone`: The zone to launch the instance.
 - `--google-machine-type`: The type of instance.
 - `--google-username`: The username to use for the instance.
 - `--google-scopes`: The scopes for OAuth 2.0 to Access Google APIs. See [Google Compute Engine Doc](https://cloud.google.com/storage/docs/authentication).
 - `--google-disk-size`: The disk size of instance.
 - `--google-disk-type`: The disk type of instance.
 - `--google-address`: Instance's static external IP (name or IP).
 - `--google-preemptible`: Instance preemptibility.
 - `--google-tags`: Instance tags (comma-separated).

The driver uses the `ubuntu-1404-trusty-v20150909a` disk image.

Environment variables and default values:

| CLI option                | Environment variable  | Default                              |
|---------------------------|-----------------------|--------------------------------------|
| **`--google-project`**    | `GOOGLE_PROJECT`      | -                                    |
| `--google-zone`           | `GOOGLE_ZONE`         | `us-central1-a`                      |
| `--google-machine-type`   | `GOOGLE_MACHINE_TYPE` | `n1-standard-1`                      |
| `--google-username`       | `GOOGLE_USERNAME`     | `docker-user`                        |
| `--google-scopes`         | `GOOGLE_SCOPES`       | `devstorage.read_only,logging.write` |
| `--google-disk-size`      | `GOOGLE_DISK_SIZE`    | `10`                                 |
| `--google-disk-type`      | `GOOGLE_DISK_TYPE`    | `pd-standard`                        |
| `--google-address`        | `GOOGLE_ADDRESS`      | -                                    |
| `--google-preemptible`    | `GOOGLE_PREEMPTIBLE`  | -                                    |
| `--google-tags`           | `GOOGLE_TAGS`         | -                                    |

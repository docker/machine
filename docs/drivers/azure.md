<!--[metadata]>
+++
title = "Microsoft Azure"
description = "Microsoft Azure driver for machine"
keywords = ["machine, Microsoft Azure, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Microsoft Azure
Create machines on [Microsoft Azure](http://azure.microsoft.com/).

You need to create a subscription with a cert. Run these commands and answer the questions:

    $ openssl req -x509 -nodes -days 365 -newkey rsa:1024 -keyout mycert.pem -out mycert.pem
    $ openssl pkcs12 -export -out mycert.pfx -in mycert.pem -name "My Certificate"
    $ openssl x509 -inform pem -in mycert.pem -outform der -out mycert.cer

Go to the Azure portal, go to the "Settings" page (you can find the link at the bottom of the
left sidebar - you need to scroll), then "Management Certificates" and upload `mycert.cer`.

Grab your subscription ID from the portal, then run `docker-machine create` with these details:

    $ docker-machine create -d azure --azure-subscription-id="SUB_ID" --azure-subscription-cert="mycert.pem" A-VERY-UNIQUE-NAME

The Azure driver uses the `b39f27a8b8c64d52b05eac6a62ebad85__Ubuntu-14_04_1-LTS-amd64-server-20140927-en-us-30GB`
image by default. Note, this image is not available in the Chinese regions. In China you should
 specify `b549f4301d0b4295b8e76ceb65df47d4__Ubuntu-14_04_1-LTS-amd64-server-20140927-en-us-30GB`.

You may need to `machine ssh` in to the virtual machine and reboot to ensure that the OS is updated.

Options:

 - `--azure-docker-port`: Port for Docker daemon.
 - `--azure-image`: Azure image name. See [How to: Get the Windows Azure Image Name](https://msdn.microsoft.com/en-us/library/dn135249%28v=nav.70%29.aspx)
 - `--azure-location`: Machine instance location.
 - `--azure-password`: Your Azure password.
 - `--azure-publish-settings-file`: Azure setting file. See [How to: Download and Import Publish Settings and Subscription Information](https://msdn.microsoft.com/en-us/library/dn385850%28v=nav.70%29.aspx)
 - `--azure-size`: Azure disk size.
 - `--azure-ssh-port`: Azure SSH port.
 - `--azure-subscription-id`: **required** Your Azure subscription ID (A GUID like `d255d8d7-5af0-4f5c-8a3e-1545044b861e`).
 - `--azure-subscription-cert`: **required** Your Azure subscription cert.
 - `--azure-username`: Azure login user name.

Environment variables and default values:

| CLI option                      | Environment variable          | Default               |
|---------------------------------|-------------------------------| ----------------------|
| `--azure-docker-port`           | -                             | `2376`                |
| `--azure-image`                 | `AZURE_IMAGE`                 | *Ubuntu 14.04 LTS x64*|
| `--azure-location`              | `AZURE_LOCATION`              | `West US`             |
| `--azure-password`              | -                             | -                     |
| `--azure-publish-settings-file` | `AZURE_PUBLISH_SETTINGS_FILE` | -                     |
| `--azure-size`                  | `AZURE_SIZE`                  | `Small`               |
| `--azure-ssh-port`              | -                             | `22`                  |
| **`--azure-subscription-cert`** | `AZURE_SUBSCRIPTION_CERT`     | -                     |
| **`--azure-subscription-id`**   | `AZURE_SUBSCRIPTION_ID`       | -                     |
| `--azure-username`              | -                             | `ubuntu`              |
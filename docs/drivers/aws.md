<!--[metadata]>
+++
title = "Amazon Web Services"
description = "Amazon Web Services driver for machine"
keywords = ["machine, Amazon Web Services, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Amazon Web Services

Create machines on [Amazon Web Services](http://aws.amazon.com). To create machines on [Amazon Web Services](http://aws.amazon.com), you must supply three required parameters:

-   Access Key ID
-   Secret Access Key
-   VPC ID

Obtain your IDs and Keys from AWS. To find the VPC ID:

1.  Login to the AWS console
2.  Go to **Services -> VPC -> Your VPCs**.
3.  Locate the VPC ID you want from the _VPC_ column.
4.  Go to **Services -> VPC -> Subnets**. Examine the _Availability Zone_ column to verify that zone `a` exists and matches your VPC ID.

    For example, `us-east1-a` is in the `a` availability zone. If the `a` zone is not present, you can create a new subnet in that zone or specify a different zone when you create the machine.

To create the machine instance, specify `--driver amazonec2` and the three required parameters.

    $ docker-machine create --driver amazonec2 --amazonec2-access-key AKI******* --amazonec2-secret-key 8T93C********* --amazonec2-vpc-id vpc-****** aws01

This example assumes the VPC ID was found in the `a` availability zone. Use the`--amazonec2-zone` flag to specify a zone other than the `a` zone. For example, `--amazonec2-zone c` signifies `us-east1-c`.

## Options

-   `--amazonec2-access-key`: **required** Your access key id for the Amazon Web Services API.
-   `--amazonec2-secret-key`: **required** Your secret access key for the Amazon Web Services API.
-   `--amazonec2-session-token`: Your session token for the Amazon Web Services API.
-   `--amazonec2-ami`: The AMI ID of the instance to use.
-   `--amazonec2-region`: The region to use when launching the instance.
-   `--amazonec2-vpc-id`: **required** Your VPC ID to launch the instance in.
-   `--amazonec2-zone`: The AWS zone to launch the instance in (i.e. one of a,b,c,d,e).
-   `--amazonec2-subnet-id`: AWS VPC subnet id.
-   `--amazonec2-security-group`: AWS VPC security group name.
-   `--amazonec2-tags`: AWS extra tag key-value pairs (comma-separated, e.g. key1,value1,key2,value2).
-   `--amazonec2-instance-type`: The instance type to run.
-   `--amazonec2-device-name`: The root device name of the instance.
-   `--amazonec2-root-size`: The root disk size of the instance (in GB).
-   `--amazonec2-volume-type`: The Amazon EBS volume type to be attached to the instance.
-   `--amazonec2-iam-instance-profile`: The AWS IAM role name to be used as the instance profile.
-   `--amazonec2-ssh-user`: SSH Login user name.
-   `--amazonec2-request-spot-instance`: Use spot instances.
-   `--amazonec2-spot-price`: Spot instance bid price (in dollars). Require the `--amazonec2-request-spot-instance` flag.
-   `--amazonec2-use-private-address`: Use the private IP address for docker-machine, but still create a public IP address.
-   `--amazonec2-private-address-only`: Use the private IP address only.
-   `--amazonec2-monitoring`: Enable CloudWatch Monitoring.
-   `--amazonec2-use-ebs-optimized-instance`: Create an EBS Optimized Instance, instance type must support it.
-   `--amazonec2-ssh-keypath`: Path to Private Key file to use for instance. Matching public key with .pub extension should exist.

By default, the Amazon EC2 driver will use a daily image of Ubuntu 15.10 LTS.

| Region         | AMI ID       |
| -------------- | ------------ |
| ap-northeast-1 | ami-b36d4edd |
| ap-southeast-1 | ami-1069af73 |
| ap-southeast-2 | ami-1d336a7e |
| cn-north-1     | ami-79eb2214 |
| eu-west-1      | ami-8aa67cf9 |
| eu-central-1   | ami-ab0210c7 |
| sa-east-1      | ami-185de774 |
| us-east-1      | ami-26d5af4c |
| us-west-1      | ami-9cbcd2fc |
| us-west-2      | ami-16b1a077 |
| us-gov-west-1  | ami-b0bad893 |

Environment variables and default values:

| CLI option                               | Environment variable    | Default          |
| ---------------------------------------- | ----------------------- | ---------------- |
| **`--amazonec2-access-key`**             | `AWS_ACCESS_KEY_ID`     | -                |
| **`--amazonec2-secret-key`**             | `AWS_SECRET_ACCESS_KEY` | -                |
| `--amazonec2-session-token`              | `AWS_SESSION_TOKEN`     | -                |
| `--amazonec2-ami`                        | `AWS_AMI`               | `ami-5f709f34`   |
| `--amazonec2-region`                     | `AWS_DEFAULT_REGION`    | `us-east-1`      |
| **`--amazonec2-vpc-id`**                 | `AWS_VPC_ID`            | -                |
| `--amazonec2-zone`                       | `AWS_ZONE`              | `a`              |
| `--amazonec2-subnet-id`                  | `AWS_SUBNET_ID`         | -                |
| `--amazonec2-security-group`             | `AWS_SECURITY_GROUP`    | `docker-machine` |
| `--amazonec2-tags`                       | `AWS_TAGS`              | -                |
| `--amazonec2-instance-type`              | `AWS_INSTANCE_TYPE`     | `t2.micro`       |
| `--amazonec2-device-name`                | `AWS_DEVICE_NAME`       | `/dev/sda1`      |
| `--amazonec2-root-size`                  | `AWS_ROOT_SIZE`         | `16`             |
| `--amazonec2-volume-type`                | `AWS_VOLUME_TYPE`       | `gp2`            |
| `--amazonec2-iam-instance-profile`       | `AWS_INSTANCE_PROFILE`  | -                |
| `--amazonec2-ssh-user`                   | `AWS_SSH_USER`          | `ubuntu`         |
| `--amazonec2-request-spot-instance`      | -                       | `false`          |
| `--amazonec2-spot-price`                 | -                       | `0.50`           |
| `--amazonec2-use-private-address`        | -                       | `false`          |
| `--amazonec2-private-address-only`       | -                       | `false`          |
| `--amazonec2-monitoring`                 | -                       | `false`          |
| `--amazonec2-use-ebs-optimized-instance` | -                       | `false`          |
| `--amazonec2-ssh-keypath`                |`AWS_SSH_KEYPATH`        | -                |

## Security Group

Note that a security group will be created and associated to the host. This security group will have the following ports opened inbound :

-   ssh (22/tcp)
-   docker (2376/tcp)
-   swarm (3376/tcp), only if the node is a swarm master

If you specify a security group yourself using the `--amazonec2-security-group` flag, the above ports will be checked and opened and the security group modified.
If you want more ports to be opened, like application specific ports, use the aws console and modify the configuration manually.
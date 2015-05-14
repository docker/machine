<!--[metadata]>
+++
title = "Aliyun Elastic Compute Service"
description = "Aliyun driver for machine"
keywords = ["machine, aliyun, driver, ecs"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->
# Aliyun Elastic Compute Service
Create machines on [Aliyun Elastic Compute Service (ECS)](http://www.aliyun.com/).  You will need an Access Key ID, Secret Access Key and a Region ID.  If you want to setup Docker Cluster on the VPC network, you will need the VPC ID; Please login to the Aliyun console -> Products and Services -> VPC or go to [VPC Control Panel](http://www.aliyunvpc.com/) directly and select the one where you would like to launch the instance.


Options:

 - `--aliyunecs-access-key-id`: **required** Your access key ID for the Aliyun ECS API.
 - `--aliyunecs-access-key-secret`: **required** Your secret access key for the Aliyun ECS API.
 - `--aliyunecs-image-id`: The image ID of the instance to use Default is the latest Ubuntu 14.04 provided by system
 - `--aliyunecs-instance-type`: The instance type to run.  Default: `ecs.t1.small`
 - `--aliyunecs-internet-max-bandwidth`: Maxium bandwidth for Internet access (in Mbps), default 1
 - `--aliyunecs-private-address-only`: Use the private IP address only
 - `--aliyunecs-region`: The region to use when launching the instance. Default: `cn-hangzhou`
 - `--aliyunecs-route-cidr`: The CIDR to use configure the route entry for the instance in VPC. Sample: 192.168.200.0/24
 - `--aliyunecs-security-group`: Aliyun security group name. Default: `docker-machine`
 - `--aliyunecs-ssh-password`: SSH password for created virtual machine. Default is random generated.
 - `--aliyunecs-vpc-id`: Your VPC ID to launch the instance in. (required for VPC network only)
 - `--aliyunecs-vswitch-id`: Your VSwitch ID to launch the instance with. (required for VPC network only)
 - `--aliyunecs-zone`: The availabilty zone launch the instance

Environment variables and default values:

| CLI option                          | Environment variable        | Default          |
|-------------------------------------|-----------------------------|------------------|
| **`--aliyunecs-access-key-id`**     | `ECS_ACCESS_KEY_ID`         | -                |
| **`--aliyunecs-access-key-key`**    | `ECS_ACCESS_KEY_SECRET`     | -                |
| `--aliyunecs-image-id`              | `ECS_IMAGE_ID`              | -                |
| `--aliyunecs-instance-type`         | `ECS_INSTANCE_TYPE`         | `ecs.t1.small`   |
| `--aliyunecs-internet-max-bandwidth`| `ECS_INTERNET_MAX_BANDWIDTH`| `1`              |
| `--aliyunecs-private-address-only`  | -                           | `false`          |
| `--aliyunecs-region`                | `ECS_REGION`                | `cn-hangzhou`    |
| `--aliyunecs-route-cidr`            | `ECS_ROUTE_CIDR`            | -                |
| `--aliyunecs-security-group`        | `ECS_SECURITY_GROUP`        | -                |
| `--aliyunecs-ssh-password`          | `ECS_SSH_PASSWORD`          | Random generated |
| `--aliyunecs-vpc-id`                | `ECS_VPC_ID`                | -                |
| `--aliyunecs-vswitch-id`            | `ECS_VSWITCH_ID`            | -                |
| `--aliyunecs-zone`                  | `ECS_ZONE`                  | -                |

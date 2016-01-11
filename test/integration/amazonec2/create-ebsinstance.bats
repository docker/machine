#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

force_env DRIVER amazonec2

#Use Instance Type that supports EBS Optimize 
export AWS_INSTANCE_TYPE=m4.large

only_if_env AWS_DEFAULT_REGION

only_if_env AWS_ACCESS_KEY_ID

only_if_env AWS_SECRET_ACCESS_KEY

only_if_env AWS_SUBNET_ID


@test "$DRIVER: Should Create an EBS Optimized Instance" {
  
  machine create -d amazonec2 --amazonec2-use-ebs-optimized-instance $NAME

  run docker $(machine config $NAME) run --rm -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY -e AWS_DEFAULT_REGION=$AWS_DEFAULT_REGION blendle/aws-cli ec2 describe-instances --filters Name=tag:Name,Values=$NAME Name=instance-state-name,Values=running --query 'Reservations[0].Instances[0].EbsOptimized' --output text	

    [[ ${lines[*]:-1} =~ "True" ]]

 }



 
 
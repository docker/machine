#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

force_env DRIVER amazonec2

require_env AWS_VPC_ID
require_env AWS_ACCESS_KEY_ID
require_env AWS_SECRET_ACCESS_KEY

@test "$DRIVER: create using RedHat AMI" {
  # Oh snap, recursive stuff!!
  AWS_AMI=ami-12663b7a AWS_SSH_USER=ec2-user run ${BASE_TEST_DIR}/run-bats.sh ${BASE_TEST_DIR}/core
  echo ${output}
  [ ${status} -eq 0 ]
}

#!/usr/bin/env bats

load helpers

export DRIVER=amazonec2
export NAME="bats-$DRIVER-test"
export MACHINE_STORAGE_PATH=/tmp/machine-bats-test-$DRIVER

load common-create
load common-start-stop
load common-destroy

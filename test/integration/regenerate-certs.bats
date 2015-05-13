#!/usr/bin/env bats

load helpers

export DRIVER=virtualbox
export NAME="bats-$DRIVER-test"
export MACHINE_STORAGE_PATH=/tmp/machine-bats-test-$DRIVER

@test "$DRIVER: create" {
  run machine create -d $DRIVER $NAME
}

@test "$DRIVER: regenerate the certs" {
  run machine regenerate-certs -f $NAME
  [[ ${status} -eq 0 ]]
}

@test "$DRIVER: make sure docker still works" {
  run docker $(machine config $NAME) version
  [[ ${status} -eq 0 ]]
}

@test "cleanup" {
  machine rm $NAME
  [[ ${status} -eq 0 ]]
}

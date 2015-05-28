#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "$DRIVER: create" {
  machine create -d $DRIVER $NAME
}

@test "Should not be able to remove running VM" {
  machine rm $NAME
  [ ${status} -eq 1 ]
}

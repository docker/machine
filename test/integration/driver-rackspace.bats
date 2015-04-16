#!/usr/bin/env bats

load helpers

export DRIVER=rackspace
export NAME="bats-$DRIVER-test"
export MACHINE_STORAGE_PATH=/tmp/machine-bats-test-$DRIVER

load common-create

@test "$DRIVER: stop should fail (unsupported)" {
  run machine stop $NAME
  [[ ${lines[1]} == *"not currently support"*  ]]
}

@test "$DRIVER: start should fail (unsupported)" {
  run machine start $NAME
  [[ ${lines[1]} == *"not currently support"*  ]]
}

@test "$DRIVER: restart" {
  run machine restart $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show running after restart" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

load common-destroy

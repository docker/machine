#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

force_env DRIVER parallels

@test "$DRIVER: create" {
  run machine create -d $DRIVER $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: prlctl pause" {
  run prlctl pause $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show paused after 'prlctl pause'" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Paused"*  ]]
}

@test "$DRIVER: start after paused" {
  run machine start $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show running after start" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

@test "$DRIVER: prlctl suspend" {
  run prlctl suspend $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show saved after 'prlctl suspend'" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"$NAME"*  ]]
  [[ ${lines[1]} == *"Saved"*  ]]
}

@test "$DRIVER: start after saved" {
  run machine start $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show running after start" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

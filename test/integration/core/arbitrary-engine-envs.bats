#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash


@test "$DRIVER: create with arbitrary engine envs" {

run machine create -d $DRIVER \
  --engine-env=TEST=VALUE \
  $NAME
  [ $status -eq 0  ]
}

@test "$DRIVER: test docker process envs" {

  # get pid of docker process, check process envs for set Environment Variable from above test
  run machine ssh $NAME 'pgrep -f "docker -d" | xargs -I % sudo cat /proc/%/environ | grep -q "TEST=VALUE"'
  [ $status -eq 0 ]
}

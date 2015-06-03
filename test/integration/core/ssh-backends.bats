#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

# Basic smoke test for SSH backends

@test "$DRIVER: create SSH test box" {
  run machine create -d $DRIVER $NAME
  [[ "$status" -eq 0  ]]
}

@test "$DRIVER: test external ssh backend" {
  run machine ssh $NAME -- df -h
  [[ "$status" -eq 0 ]]
}

@test "$DRIVER: test command did what it purported to -- external ssh" {
  run machine ssh $NAME echo foo
  [[ "$output" == "foo"  ]]
}

@test "$DRIVER: test native ssh backend" {
  run machine --native-ssh ssh $NAME -- df -h
  [[ "$status" -eq 0  ]]
}

@test "$DRIVER: test command did what it purported to -- native ssh" {
  run machine --native-ssh ssh $NAME echo foo
  [[ "$output" == "foo"  ]]
}

#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

only_if_env DRIVER ci-test

@test "create" {
  run machine create -d $DRIVER --url none default
  [ "$status" -eq 0 ]
}

@test "ls" {
  run machine ls -q
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 1 ]]
  [[ ${lines[0]} = "default" ]]
}

@test "url" {
  run machine url default
  [ "$status" -eq 0 ]
  [[ ${output} == *"none"* ]]
}

@test "status" {
  run machine status default
  [ "$status" -eq 0 ]
  [[ ${output} == *"Running"* ]]
}

@test "rm" {
  run machine rm -y default
  [ "$status" -eq 0 ]
}

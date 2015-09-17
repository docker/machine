#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "status: show error in case of no args" {
  run machine inspect
  [ "$status" -eq 1 ]
  [[ ${output} == *"must specify a machine name"* ]]
}

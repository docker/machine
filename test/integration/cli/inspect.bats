#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "inspect: show error in case of no args" {
  run machine inspect
  [ "$status" -eq 1 ]
  [[ ${output} == *"Expected one machine name as an argument"* ]] || false
}

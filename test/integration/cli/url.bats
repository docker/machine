#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "url: show error in case of no args" {
  run machine url
  [ "$status" -eq 1 ]
  [[ ${output} == *"Expected one machine name as an argument."* ]]
}

#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "url: show error in case of no args" {
  run machine url
  [ "$status" -eq 1 ]
  [[ ${output} == *"Error: No machine name(s) specified and no \"default\" machine exists."* ]]
}

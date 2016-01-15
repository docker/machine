#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

only_if_env DRIVER virtualbox

use_disposable_machine

export BAD_URL="http://dev.null:9111/bad.iso"

@test "$DRIVER: Should not allow machine creation with bad ISO" {
  run machine create -d virtualbox --virtualbox-boot2docker-url $BAD_URL $NAME
  [[ ${status} -eq 1 ]]
}

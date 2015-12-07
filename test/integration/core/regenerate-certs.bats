#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

if [[ -z "$NAME" ]]; then
  export NAME="$(shared_machine_name)"
fi

@test "$DRIVER: regenerate the certs" {
  run machine regenerate-certs -f $NAME
  [[ ${status} -eq 0 ]]
}

@test "$DRIVER: make sure docker still works" {
  run docker $(machine config $NAME) version
  [[ ${status} -eq 0 ]]
}

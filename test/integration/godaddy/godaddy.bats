#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

only_if_env DRIVER godaddy

use_disposable_machine

require_env GODADDY_API_KEY

@test "$DRIVER: Should Create a default host" {
    run machine create -d godaddy $NAME
    echo ${output}
    [ "$status" -eq 0 ]
}

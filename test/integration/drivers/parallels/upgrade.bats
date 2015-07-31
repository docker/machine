#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

force_env DRIVER parallels

export OLD_ISO_URL="https://github.com/Parallels/boot2docker/releases/download/v1.7.0-prl-tools/boot2docker.iso"

@test "$DRIVER: create for upgrade" {
  run machine create -d parallels --parallels-boot2docker-url $OLD_ISO_URL $NAME
}

@test "$DRIVER: verify that docker version is old" {
  # Have to run this over SSH due to client/server mismatch restriction
  SERVER_VERSION=$(machine ssh $NAME docker version | grep 'Server version' | awk '{ print $3; }')
  [[ "$SERVER_VERSION" == "1.7.0" ]]
}

@test "$DRIVER: upgrade" {
  run machine upgrade $NAME
  echo ${output}
  [ "$status" -eq 0  ]
}

@test "$DRIVER: upgrade is correct version" {
  SERVER_VERSION=$(docker $(machine config $NAME) version | grep 'Server version' | awk '{ print $3; }')
  [[ "$SERVER_VERSION" != "1.7.0" ]]
}

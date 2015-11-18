#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

force_env DRIVER virtualbox

export OLD_ISO_URL="https://github.com/boot2docker/boot2docker/releases/download/v1.8.3/boot2docker.iso"

@test "$DRIVER: create for upgrade" {
  run machine create -d virtualbox --virtualbox-boot2docker-url $OLD_ISO_URL $NAME
}

@test "$DRIVER: verify that docker version is old" {
  # Have to run this over SSH due to client/server mismatch restriction
  SERVER_VERSION=$(machine ssh $NAME docker version --format {{.Server.Version}})
  [ "$SERVER_VERSION" == "1.8.3" ]
}

@test "$DRIVER: upgrade" {
  run machine upgrade $NAME
  echo ${output}
  [ "$status" -eq 0  ]
}

@test "$DRIVER: upgrade is correct version" {
  NEW_SERVER_VERSION=$(machine ssh $NAME docker version --format {{.Server.Version}})
  [ "$NEW_SERVER_VERSION" != "1.8.3" ]
}

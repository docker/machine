#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "$DRIVER: create" {
  run machine create -d $DRIVER $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: test powershell notation" {
  run machine env --shell powershell $NAME
  [[ ${lines[0]} == "\$Env:DOCKER_TLS_VERIFY = \"1\"" ]]
  [[ ${lines[1]} == "\$Env:DOCKER_HOST = \"$(machine url $NAME)\"" ]]
  [[ ${lines[2]} == "\$Env:DOCKER_CERT_PATH = \"$MACHINE_STORAGE_PATH/certs\"" ]]
  [[ ${lines[3]} == "\$Env:DOCKER_MACHINE_NAME = \"$NAME\"" ]]
}

@test "$DRIVER: test bash / zsh notation" {
  run machine env $NAME
  [[ ${lines[0]} == "export DOCKER_TLS_VERIFY=\"1\"" ]]
  [[ ${lines[1]} == "export DOCKER_HOST=\"$(machine url $NAME)\"" ]]
  [[ ${lines[2]} == "export DOCKER_CERT_PATH=\"$MACHINE_STORAGE_PATH/certs\"" ]]
  [[ ${lines[3]} == "export DOCKER_MACHINE_NAME=\"$NAME\"" ]]
}

@test "$DRIVER: test cmd.exe notation" {
  run machine env --shell cmd $NAME
  [[ ${lines[0]} == "set DOCKER_TLS_VERIFY=1" ]]
  [[ ${lines[1]} == "set DOCKER_HOST=$(machine url $NAME)" ]]
  [[ ${lines[2]} == "set DOCKER_CERT_PATH=$MACHINE_STORAGE_PATH/certs" ]]
  [[ ${lines[3]} == "set DOCKER_MACHINE_NAME=$NAME" ]]
}

@test "$DRIVER: test fish notation" {
  run machine env --shell fish $NAME
  [[ ${lines[0]} == "set -x DOCKER_TLS_VERIFY \"1\";" ]]
  [[ ${lines[1]} == "set -x DOCKER_HOST \"$(machine url $NAME)\";" ]]
  [[ ${lines[2]} == "set -x DOCKER_CERT_PATH \"$MACHINE_STORAGE_PATH/certs\";" ]]
  [[ ${lines[3]} == "set -x DOCKER_MACHINE_NAME \"$NAME\";" ]]
}

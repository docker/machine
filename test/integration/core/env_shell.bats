#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "$DRIVER: create" {
  run machine create -d $DRIVER $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: test powershell notation" {
  run machine env --shell powershell --no-proxy $NAME
  [[ ${lines[0]} == "\$Env:DOCKER_TLS_VERIFY = \"1\"" ]]
  [[ ${lines[1]} == "\$Env:DOCKER_HOST = \"$(machine url $NAME)\"" ]]
  [[ ${lines[2]} == "\$Env:DOCKER_CERT_PATH = \"$MACHINE_STORAGE_PATH/machines/$NAME\"" ]]
  [[ ${lines[3]} == "\$Env:DOCKER_MACHINE_NAME = \"$NAME\"" ]]
  [[ ${lines[4]} == "\$Env:NO_PROXY = \"$(machine ip $NAME)\"" ]]
}

@test "$DRIVER: test bash / zsh notation" {
  run machine env --no-proxy $NAME
  [[ ${lines[0]} == "export DOCKER_TLS_VERIFY=\"1\"" ]]
  [[ ${lines[1]} == "export DOCKER_HOST=\"$(machine url $NAME)\"" ]]
  [[ ${lines[2]} == "export DOCKER_CERT_PATH=\"$MACHINE_STORAGE_PATH/machines/$NAME\"" ]]
  [[ ${lines[3]} == "export DOCKER_MACHINE_NAME=\"$NAME\"" ]]
  [[ ${lines[4]} == "export NO_PROXY=\"$(machine ip $NAME)\"" ]]
}

@test "$DRIVER: test cmd.exe notation" {
  run machine env --shell cmd --no-proxy $NAME
  [[ ${lines[0]} == "SET DOCKER_TLS_VERIFY=1" ]]
  [[ ${lines[1]} == "SET DOCKER_HOST=$(machine url $NAME)" ]]
  [[ ${lines[2]} == "SET DOCKER_CERT_PATH=$MACHINE_STORAGE_PATH/machines/$NAME" ]]
  [[ ${lines[3]} == "SET DOCKER_MACHINE_NAME=$NAME" ]]
  [[ ${lines[4]} == "SET NO_PROXY=$(machine ip $NAME)" ]]
}

@test "$DRIVER: test fish notation" {
  run machine env --shell fish --no-proxy $NAME
  [[ ${lines[0]} == "set -x DOCKER_TLS_VERIFY \"1\";" ]]
  [[ ${lines[1]} == "set -x DOCKER_HOST \"$(machine url $NAME)\";" ]]
  [[ ${lines[2]} == "set -x DOCKER_CERT_PATH \"$MACHINE_STORAGE_PATH/machines/$NAME\";" ]]
  [[ ${lines[3]} == "set -x DOCKER_MACHINE_NAME \"$NAME\";" ]]
  [[ ${lines[4]} == "set -x NO_PROXY \"$(machine ip $NAME)\";" ]]
}

@test "$DRIVER: no proxy with NO_PROXY already set" {
  export NO_PROXY=localhost
  run machine env --no-proxy $NAME
  [[ ${lines[4]} == "export NO_PROXY=\"localhost,$(machine ip $NAME)\"" ]]
}

@test "$DRIVER: test powershell unset env" {
  run machine env --shell powershell --no-proxy --unset
  [[ ${lines[0]} == "Remove-Item Env:\\\\DOCKER_TLS_VERIFY" ]]
  [[ ${lines[1]} == "Remove-Item Env:\\\\DOCKER_HOST" ]]
  [[ ${lines[2]} == "Remove-Item Env:\\\\DOCKER_CERT_PATH" ]]
  [[ ${lines[3]} == "Remove-Item Env:\\\\DOCKER_MACHINE_NAME" ]]
  [[ ${lines[4]} == "Remove-Item Env:\\\\NO_PROXY" ]]
}

@test "$DRIVER: test bash / zsh unset env" {
  run machine env --no-proxy --unset
  [[ ${lines[0]} == "unset DOCKER_TLS_VERIFY" ]]
  [[ ${lines[1]} == "unset DOCKER_HOST" ]]
  [[ ${lines[2]} == "unset DOCKER_CERT_PATH" ]]
  [[ ${lines[3]} == "unset DOCKER_MACHINE_NAME" ]]
  [[ ${lines[4]} == "unset NO_PROXY" ]]
}

@test "$DRIVER: test cmd.exe unset env" {
  run machine env --shell cmd --no-proxy --unset
  [[ ${lines[0]} == "SET DOCKER_TLS_VERIFY=" ]]
  [[ ${lines[1]} == "SET DOCKER_HOST=" ]]
  [[ ${lines[2]} == "SET DOCKER_CERT_PATH=" ]]
  [[ ${lines[3]} == "SET DOCKER_MACHINE_NAME=" ]]
  [[ ${lines[4]} == "SET NO_PROXY=" ]]
}

@test "$DRIVER: test fish unset env" {
  run machine env --shell fish --no-proxy --unset
  [[ ${lines[0]} == "set -e DOCKER_TLS_VERIFY;" ]]
  [[ ${lines[1]} == "set -e DOCKER_HOST;" ]]
  [[ ${lines[2]} == "set -e DOCKER_CERT_PATH;" ]]
  [[ ${lines[3]} == "set -e DOCKER_MACHINE_NAME;" ]]
  [[ ${lines[4]} == "set -e NO_PROXY;" ]]
}

@test "$DRIVER: one arg with --unset flag fails" {
  run machine env --unset $NAME
  echo ${output}
  [ "$status" -eq 1  ]
  [[ ${lines[0]} == "Error: Expected either one machine name, or -u flag to unset the variables in the arguments" ]]
}

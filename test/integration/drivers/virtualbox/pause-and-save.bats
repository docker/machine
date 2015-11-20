#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

## THIS IS VIRTUALBOX ONLY

force_env DRIVER virtualbox

@test "$DRIVER: create a new virtualbox machine" {
  run machine create -d $DRIVER $NAME
  echo ${output}
  [ "$status" -eq 0  ]
}

@test "$DRIVER: pause the newly created machine" {
  run vboxmanage controlvm $NAME pause
  echo ${output}
  [ "$status" -eq 0 ]
}

@test "$DRIVER: status should show paused after pause" {
  run machine status $NAME
  echo ${output}
  [ "$status" -eq 0 ]
  [[ ${output} == *"Paused"* ]]
}

@test "$DRIVER: should stop a paused machine" {
  run machine stop $NAME
  echo ${output}
  [ "$status" -eq 0 ]
}

@test "$DRIVER: status should show Stopped after stop" {
  run machine status $NAME
  echo ${output}
  [ "$status" -eq 0 ]
  [[ ${output} == *"Stopped"* ]]
}

@test "$DRIVER: restart the machine" {
  run machine start $NAME
  echo ${output}
  [ "$status" -eq 0 ]
}

@test "$DRIVER: status should show Running after restart" {
  run machine status $NAME
  echo ${output}
  [ "$status" -eq 0 ]
  [[ ${output} == *"Running"* ]]
}

@test "$DRIVER: savestate the machine" {
  run VBoxManage controlvm $NAME savestate
  [ "$status" -eq 0  ]
}

@test "$DRIVER: status should show Saved after save" {
  run machine status $NAME
  echo ${output}
  [ "$status" -eq 0 ]
  [[ ${output} == *"Saved"* ]]
}

@test "$DRIVER: should start after save" {
  run machine start $NAME
  echo ${output}
  [ "$status" -eq 0 ]
}

@test "$DRIVER: status should show Running after restart" {
  run machine status $NAME
  echo ${output}
  [ "$status" -eq 0 ]
  [[ ${output} == *"Running"* ]]
}

@test "$DRIVER: pause the machine again" {
  run vboxmanage controlvm $NAME pause
  echo ${output}
  [ "$status" -eq 0 ]
}

@test "$DRIVER: remove the paused machine" {
  run machine rm $NAME
  echo ${output}
  [ "$status" -eq 0 ]
}
#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "$DRIVER: machine should not exist" {
  run machine inspect $NAME
  echo ${output}
  [ "$status" -eq 1  ]
}

@test "$DRIVER: create" {
  run machine create -d $DRIVER $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: ls" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"$NAME"*  ]]
}

@test "$DRIVER: run busybox container" {
  run docker $(machine config $NAME) run busybox echo hello world
  [ "$status" -eq 0  ]
}

@test "$DRIVER: url" {
  run machine url $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: ip" {
  run machine ip $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: ssh" {
  run machine ssh $NAME -- ls -lah /
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "total"  ]]
}

@test "$DRIVER: docker commands with the socket should work" {
  run machine ssh $NAME -- docker version
}

@test "$DRIVER: stop" {
  run machine stop $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show stopped after stop" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Stopped"*  ]]
}

@test "$DRIVER: url should show an error when machine is stopped" {
  run machine url $NAME
  [ "$status" -eq 1 ]
  [[ ${output} == *"not running"* ]]
}

@test "$DRIVER: env should show an error when machine is stopped" {
  run machine env $NAME
  [ "$status" -eq 1 ]
  [[ ${output} == *"not running. Please start this with"* ]]
}

@test "$DRIVER: machine should not allow upgrade when stopped" {
  run machine upgrade $NAME
  [[ "$status" -eq 1 ]]
}

@test "$DRIVER: start" {
  run machine start $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show running after start" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

@test "$DRIVER: kill" {
  run machine kill $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show stopped after kill" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Stopped"*  ]]
}

@test "$DRIVER: restart" {
  run machine restart $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show running after restart" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

@test "$DRIVER: status" {
  run machine status $NAME
  [ "$status" -eq 0 ]
  [[ ${output} == *"Running"* ]]
}

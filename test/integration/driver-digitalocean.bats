#!/usr/bin/env bats

load helpers

export DRIVER=digitalocean
export NAME="bats-$DRIVER-test"
export MACHINE_STORAGE_PATH=/tmp/machine-bats-test-$DRIVER

@test "$DRIVER: machine should not exist" {
  run machine active $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: create" {
  run machine create -d $DRIVER $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: active" {
  run machine active $NAME
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

@test "$DRIVER: remove" {
  run sleep 20
  run machine rm -f $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should not exist" {
  run machine active $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: cleanup" {
  run rm -rf $MACHINE_STORAGE_PATH
  [ "$status" -eq 0  ]
}


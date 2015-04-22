#!/usr/bin/env bats

load helpers

export DRIVER=vmwarefusion
export NAME="bats-$DRIVER-test"
export MACHINE_STORAGE_PATH=/tmp/docker-machine-bats-test-$DRIVER

function setup() {
  # add sleep because vbox; ugh
  sleep 1
}

@test "$DRIVER: docker-machine should not exist" {
  run docker-machine active $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: create" {
  run docker-machine create -d $DRIVER $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: active" {
  run docker-machine active $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: ls" {
  run docker-machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"$NAME"*  ]]
}

@test "$DRIVER: run busybox container" {
  run docker $(docker-machine config $NAME) run busybox echo hello world
  [ "$status" -eq 0  ]
}

@test "$DRIVER: url" {
  run docker-machine url $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: ip" {
  run docker-machine ip $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: ssh" {
  run docker-machine ssh $NAME -- ls -lah /
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "total"  ]]
}

@test "$DRIVER: stop" {
  run docker-machine stop $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: docker-machine should show stopped after stop" {
  run docker-machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Stopped"*  ]]
}

@test "$DRIVER: start" {
  run docker-machine start $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: docker-machine should show running after start" {
  run docker-machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

@test "$DRIVER: kill" {
  run docker-machine kill $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: docker-machine should show stopped after kill" {
  run docker-machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Stopped"*  ]]
}

@test "$DRIVER: restart" {
  run docker-machine restart $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: docker-machine should show running after restart" {
  run docker-machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

@test "$DRIVER: remove" {
  run docker-machine rm -f $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: docker-machine should not exist" {
  run docker-machine active $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: cleanup" {
  run rm -rf $MACHINE_STORAGE_PATH
  [ "$status" -eq 0  ]
}

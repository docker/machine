#!/usr/bin/env bats

load vars

DRIVER=virtualbox

NAME="bats-$DRIVER-test"
MACHINE_STORAGE_PATH=/tmp/machine-bats-test-$DRIVER

@test "$DRIVER: machine should not exist" {
  run ./docker-machine_$PLATFORM-$ARCH active $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: create" {
  run ./docker-machine_$PLATFORM-$ARCH create -d $DRIVER $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: active" {
  run ./docker-machine_$PLATFORM-$ARCH active $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: ls" {
  run ./docker-machine_$PLATFORM-$ARCH ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} =~ "$NAME"  ]]
  [[ ${lines[1]} =~ "*"  ]]
}

@test "$DRIVER: url" {
  run ./docker-machine_$PLATFORM-$ARCH url $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: ip" {
  run ./docker-machine_$PLATFORM-$ARCH ip $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: ssh" {
  run ./docker-machine_$PLATFORM-$ARCH ssh $NAME -- ls -lah /
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "total"  ]]
}

@test "$DRIVER: stop" {
  run ./docker-machine_$PLATFORM-$ARCH stop $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show stopped after stop" {
  run ./docker-machine_$PLATFORM-$ARCH ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} =~ "$NAME"  ]]
  [[ ${lines[1]} =~ "Stopped"  ]]
}

@test "$DRIVER: start" {
  run ./docker-machine_$PLATFORM-$ARCH start $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show running after start" {
  run ./docker-machine_$PLATFORM-$ARCH ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} =~ "$NAME"  ]]
  [[ ${lines[1]} =~ "Running"  ]]
}

@test "$DRIVER: kill" {
  run ./docker-machine_$PLATFORM-$ARCH kill $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show stopped after kill" {
  run ./docker-machine_$PLATFORM-$ARCH ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} =~ "$NAME"  ]]
  [[ ${lines[1]} =~ "Stopped"  ]]
}

@test "$DRIVER: restart" {
  run ./docker-machine_$PLATFORM-$ARCH restart $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show running after restart" {
  run ./docker-machine_$PLATFORM-$ARCH ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} =~ "$NAME"  ]]
  [[ ${lines[1]} =~ "Running"  ]]
}

@test "$DRIVER: remove" {
  run ./docker-machine_$PLATFORM-$ARCH rm -f $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should not exist" {
  run ./docker-machine_$PLATFORM-$ARCH active $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: cleanup" {
  run rm -rf $MACHINE_STORAGE_PATH
  [ "$status" -eq 0  ]
}


#!/usr/bin/env bats

load helpers

export DRIVER=digitalocean
export NAME="bats-$DRIVER-test"
export MACHINE_STORAGE_PATH=/tmp/machine-bats-test-$DRIVER

@test "$DRIVER: machine should not exist" {
  run machine inspect $NAME
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

# currently this only checks debian/ubuntu style using 127.0.0.1
@test "$DRIVER: hostname should be set properly" {
  run machine ssh $NAME -- "grep -Fxq '127.0.1.1 $NAME' /etc/hosts"
  [ "$status" -eq 0  ]
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

@test "$DRIVER: create with arbitrary engine option" {
  run machine create -d $DRIVER \
    --engine-flag log-driver=none \
    $NAME
  [ $status -eq 0 ]
}

@test "$DRIVER: check created engine option (log driver)" {
  docker $(machine config $NAME) run --name nolog busybox echo this should not be logged
  run docker $(machine config $NAME) logs nolog
  [ $status -eq 1 ]
}

@test "$DRIVER: rm after arbitrary engine option create" {
  run machine rm $NAME
  [ $status -eq 0 ]
}

@test "$DRIVER: create with supported engine options" {
  run machine create -d $DRIVER \
    --engine-label spam=eggs \
    --engine-storage-driver devicemapper \
    --engine-insecure-registry registry.myco.com \
    $NAME
  echo "$output"
  [ $status -eq 0 ]
}

@test "$DRIVER: check for engine labels" {
  spamlabel=$(docker $(machine config $NAME) info | grep spam)
  [[ $spamlabel =~ "spam=eggs" ]]
}

@test "$DRIVER: check for engine storage driver" {
  storage_driver_info=$(docker $(machine config $NAME) info | grep "Storage Driver")
  [[ $storage_driver_info =~ "devicemapper" ]]
}

@test "$DRIVER: check for insecure registry setting" {
  ir_option=$(machine ssh $NAME -- cat /etc/default/docker | grep insecure-registry)
  [[ $ir_option =~ "registry.myco.com" ]]
}

@test "$DRIVER: rm after supported engine option create" {
  run machine rm $NAME
  [ $status -eq 0 ]
}


@test "$DRIVER: machine should not exist" {
  run machine inspect $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: cleanup" {
  run rm -rf $MACHINE_STORAGE_PATH
  [ "$status" -eq 0  ]
}


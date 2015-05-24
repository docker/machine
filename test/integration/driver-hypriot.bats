#!/usr/bin/env bats

load helpers

export DRIVER=hypriot
export NAME="bats-$DRIVER-test"
export MACHINE_STORAGE_PATH=/tmp/machine-bats-test-$DRIVER
[ -z "$HYPRIOT_IP_ADDRESS" ] && export HYPRIOT_IP_ADDRESS="192.168.1.233"

@test "$DRIVER: machine $HYPRIOT_IP_ADDRESS is reachable" {
  run ping -c 1 -t 1 $HYPRIOT_IP_ADDRESS
  [ "$status" -eq 0 ]
}

@test "$DRIVER: machine should not exist" {
  run machine inspect $NAME
  [ "$status" -eq 1 ]
}

@test "$DRIVER: create" {
  run machine create -d $DRIVER --hypriot-ip-address=$HYPRIOT_IP_ADDRESS $NAME
  [ "$status" -eq 0 ]
}

@test "$DRIVER: ls" {
  run machine ls
  [ "$status" -eq 0 ]
  [[ ${lines[1]} == *"$NAME"* ]]
}

@test "$DRIVER: run a container" {
  run docker $(machine config $NAME) run hypriot/rpi-node echo hello world
  [ "$status" -eq 0 ]
}

@test "$DRIVER: url" {
  run machine url $NAME
  [ "$status" -eq 0 ]
}

@test "$DRIVER: ip" {
  run machine ip $NAME
  [ "$status" -eq 0 ]
}

@test "$DRIVER: ssh" {
  run machine ssh $NAME -- ls -lah /
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "total" ]]
}

@test "$DRIVER: stop is not supported" {
  run machine stop $NAME
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "does not support" ]]
}

@test "$DRIVER: machine should show running after stop" {
  run machine ls
  [ "$status" -eq 0 ]
  [[ ${lines[1]} == *"Running"* ]]
}

@test "$DRIVER: start not supported" {
  run machine start $NAME
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "does not support" ]]
}

@test "$DRIVER: machine should show running after start" {
  run machine ls
  [ "$status" -eq 0 ]
  [[ ${lines[1]} == *"Running"* ]]
}

@test "$DRIVER: kill not supported" {
  run machine kill $NAME
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "does not support" ]]
}

@test "$DRIVER: machine should show running after kill" {
  run machine ls
  [ "$status" -eq 0 ]
  [[ ${lines[1]} == *"Running"* ]]
}

@test "$DRIVER: restart not supported" {
  run machine restart $NAME
  [ "$status" -eq 0 ]
}

@test "$DRIVER: machine should show running after restart" {
  run machine ls
  [ "$status" -eq 0 ]
  [[ ${lines[1]} == *"Running"* ]]
}

@test "$DRIVER: remove" {
  run machine rm -f $NAME
  [ "$status" -eq 0 ]
}

@test "$DRIVER: machine should not exist" {
  run machine inspect $NAME
  [ "$status" -eq 1 ]
}

@test "$DRIVER: cleanup" {
  run rm -rf $MACHINE_STORAGE_PATH
  [ "$status" -eq 0 ]
}


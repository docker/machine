#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

force_env DRIVER hypriot

export NAME="bats-$DRIVER-test"
export MACHINE_STORAGE_PATH=/tmp/machine-bats-test-$DRIVER
[ -z "$HYPRIOT_IP_ADDRESS" ] && export HYPRIOT_IP_ADDRESS="192.168.1.233"

@test "$DRIVER: machine ip=$HYPRIOT_IP_ADDRESS is reachable" {
  run ping -c 1 -t 1 $HYPRIOT_IP_ADDRESS
  [ "$status" -eq 0 ]
}

@test "$DRIVER: machine $NAME should not exist" {
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

@test "$DRIVER: config" {
  run machine config $NAME
  [ "$status" -eq 0 ]
}

@test "$DRIVER: inspect" {
  run machine inspect $NAME
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

@test "$DRIVER: run a container" {
  run docker $(machine config $NAME) run --rm hypriot/rpi-busybox-httpd /bin/busybox echo hello world
  [ "$status" -eq 0 ]
}

@test "$DRIVER: ssh" {
  run machine ssh $NAME -- ls -lah /
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "total" ]]
}

@test "$DRIVER: ssh 'type docker' should show command docker is installed" {
  run machine ssh $NAME -- type docker
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "docker is" ]]
}

@test "$DRIVER: ssh 'sudo service docker status' should show docker daemon is running" {
  run machine ssh $NAME -- sudo service docker status
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "Docker is running." ]]
}

@test "$DRIVER: ssh 'sudo service docker stop' should stop docker daemon" {
  run machine ssh $NAME -- sudo service docker stop
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "Stopping Docker: docker." ]]
}

@test "$DRIVER: ssh 'sudo service docker start' should start docker daemon" {
  run machine ssh $NAME -- sudo service docker start
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "Starting Docker: docker." ]]
}

@test "$DRIVER: ssh 'sudo service docker status' should show docker daemon is running" {
  run machine ssh $NAME -- sudo service docker status
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "Docker is running." ]]
}

@test "$DRIVER: run a container" {
  run docker $(machine config $NAME) run --rm hypriot/rpi-busybox-httpd /bin/busybox echo hello world
  [ "$status" -eq 0 ]
}

@test "$DRIVER: ssh 'sudo service docker restart' should restart docker daemon" {
  run machine ssh $NAME -- sudo service docker restart
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "Stopping Docker: docker." ]]
  [[ ${lines[1]} =~ "Starting Docker: docker." ]]
}

@test "$DRIVER: ssh 'sudo service docker status' should show docker daemon is running" {
  run machine ssh $NAME -- sudo service docker status
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "Docker is running." ]]
}

@test "$DRIVER: run a container" {
  run docker $(machine config $NAME) run --rm hypriot/rpi-busybox-httpd /bin/busybox echo hello world
  [ "$status" -eq 0 ]
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

@test "$DRIVER: apt source /etc/apt/sources.list.d/hypriot.list should exist" {
  run machine ssh $NAME -- ls -1 /etc/apt/sources.list.d/hypriot.list
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "/etc/apt/sources.list.d/hypriot.list" ]]
}

@test "$DRIVER: apt source should contain hypriot/wheezy/main" {
  run machine ssh $NAME -- cat /etc/apt/sources.list.d/hypriot.list
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "http://repository.hypriot.com/" ]]
  [[ ${lines[0]} =~ "wheezy" ]]
  [[ ${lines[0]} =~ "main" ]]
}

@test "$DRIVER: upgrade" {
  run machine upgrade $NAME
  [ "$status" -eq 0 ]
}

@test "$DRIVER: ssh 'sudo service docker status' should show docker daemon is running" {
  run machine ssh $NAME -- sudo service docker status
  [ "$status" -eq 0 ]
  [[ ${lines[0]} =~ "Docker is running." ]]
}

@test "$DRIVER: machine should show running after upgrade" {
  run machine ls
  [ "$status" -eq 0 ]
  [[ ${lines[1]} == *"Running"* ]]
}

@test "$DRIVER: run a container" {
  run docker $(machine config $NAME) run --rm hypriot/rpi-busybox-httpd /bin/busybox echo hello world
  [ "$status" -eq 0 ]
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

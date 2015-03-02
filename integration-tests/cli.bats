#!/usr/bin/env bats

load vars

@test "cli: show info" {
  run ./docker-machine_$PLATFORM-$ARCH
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "NAME:"  ]]
  [[ ${lines[1]} =~ "Create and manage machines running Docker"  ]]
}

@test "cli: show active help" {
  run ./docker-machine_$PLATFORM-$ARCH active -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command active"  ]]
}

@test "cli: show config help" {
  run ./docker-machine_$PLATFORM-$ARCH config -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command config"  ]]
}

@test "cli: show inspect help" {
  run ./docker-machine_$PLATFORM-$ARCH inspect -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command inspect"  ]]
}

@test "cli: show ip help" {
  run ./docker-machine_$PLATFORM-$ARCH ip -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command ip"  ]]
}

@test "cli: show kill help" {
  run ./docker-machine_$PLATFORM-$ARCH kill -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command kill"  ]]
}

@test "cli: show ls help" {
  run ./docker-machine_$PLATFORM-$ARCH ls -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command ls"  ]]
}

@test "cli: show restart help" {
  run ./docker-machine_$PLATFORM-$ARCH restart -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command restart"  ]]
}

@test "cli: show rm help" {
  run ./docker-machine_$PLATFORM-$ARCH rm -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command rm"  ]]
}

@test "cli: show env help" {
  run ./docker-machine_$PLATFORM-$ARCH env -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command env"  ]]
}

@test "cli: show ssh help" {
  run ./docker-machine_$PLATFORM-$ARCH ssh -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command ssh"  ]]
}

@test "cli: show start help" {
  run ./docker-machine_$PLATFORM-$ARCH start -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command start"  ]]
}

@test "cli: show stop help" {
  run ./docker-machine_$PLATFORM-$ARCH stop -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command stop"  ]]
}

@test "cli: show upgrade help" {
  run ./docker-machine_$PLATFORM-$ARCH upgrade -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command upgrade"  ]]
}

@test "cli: show url help" {
  run ./docker-machine_$PLATFORM-$ARCH url -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command url"  ]]
}

@test "flag: show version" {
  run ./docker-machine_$PLATFORM-$ARCH -v
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "version"  ]]
}

@test "flag: show help" {
  run ./docker-machine_$PLATFORM-$ARCH --help
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "NAME"  ]]
}

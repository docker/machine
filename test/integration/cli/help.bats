#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "cli: show info" {
  run machine
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "Usage:"  ]]
  [[ ${lines[1]} =~ "Create and manage machines running Docker"  ]]
}

@test "cli: show active help" {
  run machine active -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine active"  ]]
}

@test "cli: show config help" {
  run machine config -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine config"  ]]
}

@test "cli: show inspect help" {
  run machine inspect -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine inspect"  ]]
}

@test "cli: show ip help" {
  run machine ip -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine ip"  ]]
}

@test "cli: show kill help" {
  run machine kill -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine kill"  ]]
}

@test "cli: show ls help" {
  run machine ls -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine ls"  ]]
}

@test "cli: show restart help" {
  run machine restart -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine restart"  ]]
}

@test "cli: show rm help" {
  run machine rm -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine rm"  ]]
}

@test "cli: show env help" {
  run machine env -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine env"  ]]
}

@test "cli: show ssh help" {
  run machine ssh -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine ssh"  ]]
}

@test "cli: show start help" {
  run machine start -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine start"  ]]
}

@test "cli: show stop help" {
  run machine stop -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine stop"  ]]
}

@test "cli: show upgrade help" {
  run machine upgrade -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine upgrade"  ]]
}

@test "cli: show url help" {
  run machine url -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine url"  ]]
}

@test "flag: show version" {
  run machine -v
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "version"  ]]
}

@test "flag: show help" {
  run machine --help
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "Usage:"  ]]
}

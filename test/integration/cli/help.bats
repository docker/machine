#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "cli: show info" {
  run machine
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "Usage:"  ]] || false
  [[ ${lines[1]} =~ "Create and manage machines running Docker"  ]] || false
}

@test "cli: show active help" {
  run machine active -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine active"  ]] || false
}

@test "cli: show config help" {
  run machine config -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine config"  ]] || false
}

@test "cli: show create help" {
  run machine create -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine create"  ]] || false
}

@test "cli: show env help" {
  run machine env -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine env"  ]] || false
}

@test "cli: show inspect help" {
  run machine inspect -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine inspect"  ]] || false
}

@test "cli: show ip help" {
  run machine ip -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine ip"  ]] || false
}

@test "cli: show kill help" {
  run machine kill -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine kill"  ]] || false
}

@test "cli: show ls help" {
  run machine ls -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine ls"  ]] || false
}

@test "cli: show regenerate-certs help" {
  run machine regenerate-certs -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine regenerate-certs"  ]] || false
}

@test "cli: show restart help" {
  run machine restart -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine restart"  ]] || false
}

@test "cli: show rm help" {
  run machine rm -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine rm"  ]] || false
}

@test "cli: show scp help" {
  run machine scp -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine scp"  ]] || false
}

@test "cli: show ssh help" {
  run machine ssh -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine ssh"  ]] || false
}

@test "cli: show start help" {
  run machine start -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine start"  ]] || false
}

@test "cli: show status help" {
  run machine status -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine status"  ]] || false
}

@test "cli: show stop help" {
  run machine stop -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine stop"  ]] || false
}

@test "cli: show upgrade help" {
  run machine upgrade -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine upgrade"  ]] || false
}

@test "cli: show url help" {
  run machine url -h
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "machine url"  ]] || false
}

@test "flag: show version" {
  run machine -v
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "version"  ]] || false
}

@test "flag: show help" {
  run machine --help
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "Usage:"  ]] || false
}

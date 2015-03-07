#!/usr/bin/env bats

load helpers

@test "cli: show info" {
  run machine
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "NAME:"  ]]
  [[ ${lines[1]} =~ "Create and manage machines running Docker"  ]]
}

@test "cli: show active help" {
  run machine active -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command active"  ]]
}

@test "cli: show config help" {
  run machine config -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command config"  ]]
}

@test "cli: show inspect help" {
  run machine inspect -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command inspect"  ]]
}

@test "cli: show ip help" {
  run machine ip -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command ip"  ]]
}

@test "cli: show kill help" {
  run machine kill -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command kill"  ]]
}

@test "cli: show ls help" {
  run machine ls -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command ls"  ]]
}

@test "cli: show restart help" {
  run machine restart -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command restart"  ]]
}

@test "cli: show rm help" {
  run machine rm -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command rm"  ]]
}

@test "cli: show env help" {
  run machine env -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command env"  ]]
}

@test "cli: show ssh help" {
  run machine ssh -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command ssh"  ]]
}

@test "cli: show start help" {
  run machine start -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command start"  ]]
}

@test "cli: show stop help" {
  run machine stop -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command stop"  ]]
}

@test "cli: show upgrade help" {
  run machine upgrade -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command upgrade"  ]]
}

@test "cli: show url help" {
  run machine url -h
  [ "$status" -eq 0  ]
  [[ ${lines[3]} =~ "command url"  ]]
}

@test "flag: show version" {
  run machine -v
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "version"  ]]
}

@test "flag: show help" {
  run machine --help
  [ "$status" -eq 0  ]
  [[ ${lines[0]} =~ "NAME"  ]]
}

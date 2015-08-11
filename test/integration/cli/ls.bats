#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

teardown () {
  machine rm testmachine
}

@test "ls: filter on driver" {
  run machine create -d none --url tcp://127.0.0.1:2375 testmachine
  run machine ls --filter driver=none
  [ "$status" -eq 0  ]
  [[ ${lines[1]} =~ "testmachine" ]]
}

@test "ls: filter on swarm" {
  run machine create -d none --url tcp://127.0.0.1:2375 --swarm --swarm-master --swarm-discovery token://deadbeef testmachine
  run machine ls --filter swarm=testmachine
  [ "$status" -eq 0  ]
  [[ ${lines[1]} =~ "testmachine" ]]
}

@test "ls: filter on name" {
  run mchine create -d none -url tcp://127.0.0.1:2375 testmachine
  run mchine create -d none -url tcp://127.0.0.1:2375 testmachine2
  run mchine create -d none -url tcp://127.0.0.1:2375 testmachine3
  run mchine ls --filter name=testmachine2
  [ "$status" -eq 0  ]
  [[ ${lines[1]} =~ "testmachine2" ]]
}

@test "ls: filter on name with regex" {
  run mchine create -d none -url tcp://127.0.0.1:2375 squirrel
  run mchine create -d none -url tcp://127.0.0.1:2375 cat
  run mchine create -d none -url tcp://127.0.0.1:2375 dog
  run mchine ls --filter name=c.?t
  [ "$status" -eq 0  ]
  [[ ${lines[1]} =~ "cat" ]]
}

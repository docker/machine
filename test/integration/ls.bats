#!/usr/bin/env bats

load helpers

teardown() {
  echo "$BATS_TEST_NAME
----------
$output
----------

" >> ${BATS_LOG}
  machine rm -f testmachine
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

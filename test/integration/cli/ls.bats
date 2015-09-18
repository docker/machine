#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

teardown () {
  machine rm testmachine
  machine rm testmachine2
  machine rm testmachine3
}

@test "ls: filter on driver 'machine ls --filter driver=none'" {
  run machine create -d none --url none testmachine3
  run machine create -d none --url none testmachine2
  run machine create -d none --url none testmachine
  run machine ls --filter driver=none
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 4 ]]
  [[ ${lines[1]} =~ "testmachine" ]]
  [[ ${lines[2]} =~ "testmachine2" ]]
  [[ ${lines[3]} =~ "testmachine3" ]]
}

@test "ls: filter on driver 'machine ls -q --filter driver=none'" {
  run machine create -d none --url none testmachine3
  run machine create -d none --url none testmachine2
  run machine create -d none --url none testmachine
  run machine ls -q --filter driver=none
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 3 ]]
  [[ ${lines[0]} == "testmachine" ]]
  [[ ${lines[1]} == "testmachine2" ]]
  [[ ${lines[2]} == "testmachine3" ]]
}

@test "ls: filter on state 'machine ls --filter state=\"\"'" {
  run machine create -d none --url none testmachine3
  run machine create -d none --url none testmachine2
  run machine create -d none --url none testmachine
  run machine ls --filter state=""
  [ "$status" -eq 0  ]
  [[ ${#lines[@]} == 4 ]]
  [[ ${lines[1]} =~ "testmachine" ]]
  [[ ${lines[2]} =~ "testmachine2" ]]
  [[ ${lines[3]} =~ "testmachine3" ]]

  # TODO: have machines in that state
  run machine ls --filter state="Running"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 1 ]]
  run machine ls --filter state="Paused"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 1 ]]
  run machine ls --filter state="Saved"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 1 ]]
  run machine ls --filter state="Stopped"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 1 ]]
  run machine ls --filter state="Stopping"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 1 ]]
  run machine ls --filter state="Starting"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 1 ]]
  run machine ls --filter state="Error"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 1 ]]
}

@test "ls: filter on state 'machine ls -q --filter state=\"\"'" {
  run machine create -d none --url none testmachine3
  run machine create -d none --url none testmachine2
  run machine create -d none --url none testmachine
  run machine ls -q --filter state=""
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 3 ]]
  [[ ${lines[0]} == "testmachine" ]]
  [[ ${lines[1]} == "testmachine2" ]]
  [[ ${lines[2]} == "testmachine3" ]]
}

@test "ls: filter on name 'machine ls --filter name=\"testmachine2\"'" {
  run machine create -d none --url none testmachine3
  run machine create -d none --url none testmachine2
  run machine create -d none --url none testmachine
  run machine ls --filter name="testmachine2"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 2 ]]
  [[ ${lines[1]} =~ "testmachine2" ]]
}

@test "ls: filter on name 'machine ls -q --filter name=\"testmachine2\"'" {
  run machine create -d none --url none testmachine3
  run machine create -d none --url none testmachine2
  run machine create -d none --url none testmachine
  run machine ls -q --filter name="testmachine2"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 1 ]]
  [[ ${lines[0]} == "testmachine2" ]]
}

@test "ls: filter on name with regex 'machine ls --filter name=\"^t.*e\"'" {
  run machine create -d none --url none testmachine3
  run machine create -d none --url none testmachine2
  run machine create -d none --url none testmachine
  run machine ls --filter name="^t.*e"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 4 ]]
  [[ ${lines[1]} =~ "testmachine" ]]
  [[ ${lines[2]} =~ "testmachine2" ]]
  [[ ${lines[3]} =~ "testmachine3" ]]
}

@test "ls: filter on name with regex 'machine ls -q --filter name=\"^t.*e\"'" {
  run machine create -d none --url none testmachine3
  run machine create -d none --url none testmachine2
  run machine create -d none --url none testmachine
  run machine ls -q --filter name="^t.*e"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 3 ]]
  [[ ${lines[0]} == "testmachine" ]]
  [[ ${lines[1]} == "testmachine2" ]]
  [[ ${lines[2]} == "testmachine3" ]]
}

@test "ls: filter on swarm 'machine ls --filter swarm=testmachine3'" {
  run machine create -d none --url tcp://127.0.0.1:2375 --swarm --swarm-master --swarm-discovery token://deadbeef testmachine3
  run machine create -d none --url tcp://127.0.0.1:2375 --swarm --swarm-discovery token://deadbeef testmachine2
  run machine create -d none --url tcp://127.0.0.1:2375 --swarm --swarm-discovery token://deadbeef testmachine
  sleep 0.5
  run machine ls --filter swarm=testmachine3
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 4 ]]
  [[ ${lines[1]} =~ "testmachine" ]]
  [[ ${lines[2]} =~ "testmachine2" ]]
  [[ ${lines[3]} =~ "testmachine3" ]]
}

@test "ls: filter on swarm 'machine ls -q --filter swarm=testmachine3'" {
  run machine create -d none --url none --swarm --swarm-master --swarm-discovery token://deadbeef testmachine3
  run machine create -d none --url none --swarm --swarm-discovery token://deadbeef testmachine2
  run machine create -d none --url none --swarm --swarm-discovery token://deadbeef testmachine
  sleep 0.5
  run machine ls -q --filter swarm=testmachine3
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 3 ]]
  [[ ${lines[0]} == "testmachine" ]]
  [[ ${lines[1]} == "testmachine2" ]]
  [[ ${lines[2]} == "testmachine3" ]]
}

@test "ls: multi filter 'machine ls -q --filter swarm=testmachine3 --filter name=\"^t.*e\" --filter driver=none --filter state=\"\"'" {
  run machine create -d none --url none --swarm --swarm-master --swarm-discovery token://deadbeef testmachine3
  run machine create -d none --url none --swarm --swarm-discovery token://deadbeef testmachine2
  run machine create -d none --url none --swarm --swarm-discovery token://deadbeef testmachine
  sleep 0.5
  run machine ls -q --filter swarm=testmachine3 --filter name="^t.*e" --filter driver=none --filter state=""
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 3 ]]
  [[ ${lines[0]} == "testmachine" ]]
  [[ ${lines[1]} == "testmachine2" ]]
  [[ ${lines[2]} == "testmachine3" ]]
}

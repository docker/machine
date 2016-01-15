#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

setup () {
  machine create -d none --url none --engine-label app=1 testmachine5
  machine create -d none --url none --engine-label foo=bar --engine-label app=1 testmachine4
  machine create -d none --url none testmachine3
  machine create -d none --url none testmachine2
  machine create -d none --url none testmachine
}

teardown () {
  machine rm -y $(machine ls -q)
  echo_to_log
}

bootstrap_swarm () {
  machine create -d none --url tcp://127.0.0.1:2375 --swarm --swarm-master --swarm-discovery token://deadbeef testswarm
  machine create -d none --url tcp://127.0.0.1:2375 --swarm --swarm-discovery token://deadbeef testswarm2
  machine create -d none --url tcp://127.0.0.1:2375 --swarm --swarm-discovery token://deadbeef testswarm3
}

@test "ls: filter on label 'machine ls --filter label=foo=bar'" {
  run machine ls --filter label=foo=bar
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 2 ]]
  [[ ${lines[1]} =~ "testmachine4" ]]
}

@test "ls: mutiple filters on label 'machine ls --filter label=foo=bar --filter label=app=1'" {
  run machine ls --filter label=foo=bar --filter label=app=1
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 3 ]]
  [[ ${lines[1]} =~ "testmachine4" ]]
  [[ ${lines[2]} =~ "testmachine5" ]]
}

@test "ls: non-existing filter on label 'machine ls --filter label=invalid=filter'" {
  run machine ls --filter label=invalid=filter
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 1 ]]
}

@test "ls: filter on driver 'machine ls --filter driver=none'" {
  run machine ls --filter driver=none
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 6 ]]
  [[ ${lines[1]} =~ "testmachine" ]]
  [[ ${lines[2]} =~ "testmachine2" ]]
  [[ ${lines[3]} =~ "testmachine3" ]]
  [[ ${lines[4]} =~ "testmachine4" ]]
  [[ ${lines[5]} =~ "testmachine5" ]]
}

@test "ls: filter on driver 'machine ls -q --filter driver=none'" {
  run machine ls -q --filter driver=none
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 5 ]]
  [[ ${lines[0]} == "testmachine" ]]
  [[ ${lines[1]} == "testmachine2" ]]
  [[ ${lines[2]} == "testmachine3" ]]
}

@test "ls: filter on state 'machine ls --filter state=\"Running\"'" {
  # Default state for 'none' driver is "Running"
  run machine ls --filter state="Running"
  [ "$status" -eq 0  ]
  [[ ${#lines[@]} == 6 ]]
  [[ ${lines[1]} =~ "testmachine" ]]
  [[ ${lines[2]} =~ "testmachine2" ]]
  [[ ${lines[3]} =~ "testmachine3" ]]

  # TODO: have machines in that state
  run machine ls --filter state="None"
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

@test "ls: filter on state 'machine ls -q --filter state=\"Running\"'" {
  run machine ls -q --filter state="Running"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 5 ]]
  [[ ${lines[0]} == "testmachine" ]]
  [[ ${lines[1]} == "testmachine2" ]]
  [[ ${lines[2]} == "testmachine3" ]]
}

@test "ls: filter on name 'machine ls --filter name=\"testmachine2\"'" {
  run machine ls --filter name="testmachine2"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 2 ]]
  [[ ${lines[1]} =~ "testmachine2" ]]
}

@test "ls: filter on name 'machine ls -q --filter name=\"testmachine2\"'" {
  run machine ls -q --filter name="testmachine2"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 1 ]]
  [[ ${lines[0]} == "testmachine2" ]]
}

@test "ls: filter on name with regex 'machine ls --filter name=\"^t.*e\"'" {
  run machine ls --filter name="^t.*e"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 6 ]]
  [[ ${lines[1]} =~ "testmachine" ]]
  [[ ${lines[2]} =~ "testmachine2" ]]
  [[ ${lines[3]} =~ "testmachine3" ]]
}

@test "ls: filter on name with regex 'machine ls -q --filter name=\"^t.*e\"'" {
  run machine ls -q --filter name="^t.*e"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 5 ]]
  [[ ${lines[0]} == "testmachine" ]]
  [[ ${lines[1]} == "testmachine2" ]]
  [[ ${lines[2]} == "testmachine3" ]]
}

@test "ls: filter on swarm 'machine ls --filter swarm=testswarm" {
  bootstrap_swarm
  run machine ls --filter swarm=testswarm
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 4 ]]
  [[ ${lines[1]} =~ "testswarm" ]]
  [[ ${lines[2]} =~ "testswarm2" ]]
  [[ ${lines[3]} =~ "testswarm3" ]]
}

@test "ls: filter on swarm 'machine ls -q --filter swarm=testswarm" {
  bootstrap_swarm
  run machine ls -q --filter swarm=testswarm
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 3 ]]
  [[ ${lines[0]} == "testswarm" ]]
  [[ ${lines[1]} == "testswarm2" ]]
  [[ ${lines[2]} == "testswarm3" ]]
}

@test "ls: multi filter 'machine ls -q --filter swarm=testswarm --filter name=\"^t.*e\" --filter driver=none --filter state=\"Running\"'" {
  bootstrap_swarm
  run machine ls -q --filter swarm=testswarm --filter name="^t.*e" --filter driver=none --filter state="Running"
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 3 ]]
  [[ ${lines[0]} == "testswarm" ]]
  [[ ${lines[1]} == "testswarm2" ]]
  [[ ${lines[2]} == "testswarm3" ]]
}

@test "ls: format on driver 'machine ls --format '{{ .DriverName }}'" {
  run machine ls --format '{{ .DriverName }}'
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 5 ]]
  [[ ${lines[0]} =~ "none" ]]
  [[ ${lines[1]} =~ "none" ]]
  [[ ${lines[2]} =~ "none" ]]
  [[ ${lines[3]} =~ "none" ]]
  [[ ${lines[4]} =~ "none" ]]
}


@test "ls: format on name and driver 'machine ls --format 'table {{ .Name}}: {{ .DriverName }}'" {
  run machine ls --format 'table {{ .Name}}: {{ .DriverName }}'
  [ "$status" -eq 0 ]
  [[ ${#lines[@]} == 6 ]]
  [[ ${lines[1]} =~ "testmachine: none" ]]
  [[ ${lines[2]} =~ "testmachine2: none" ]]
  [[ ${lines[3]} =~ "testmachine3: none" ]]
  [[ ${lines[4]} =~ "testmachine4: none" ]]
  [[ ${lines[5]} =~ "testmachine5: none" ]]
}


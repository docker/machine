#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash
export TOKEN=$(curl -sS -X POST "https://discovery-stage.hub.docker.com/v1/clusters")

@test "create swarm master" {
  run machine create -d $DRIVER --swarm --swarm-master --swarm-discovery "token://$TOKEN" queenbee
  echo ${output}
  [[ ${status} -eq 0 ]]
}

@test "should not show as swarm active if normal active" {
  eval "$(machine env queenbee)"
  run machine ls
  echo ${output}
  [[ ${lines[1]} != *"* (swarm)"*  ]]
}

@test "should show as swarm active" {
  eval "$(machine env --swarm queenbee)"
  run machine ls
  echo ${output}
  [[ ${lines[1]} == *"* (swarm)"*  ]]
}

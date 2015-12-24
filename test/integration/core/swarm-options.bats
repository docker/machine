#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

export TOKEN=$(curl -sS -X POST "https://discovery.hub.docker.com/v1/clusters")

@test "create swarm master" {
    run machine create -d $DRIVER --swarm --swarm-master --swarm-discovery "token://$TOKEN" --swarm-strategy binpack --swarm-opt heartbeat=5s queenbee
    echo ${output}
    [[ "$status" -eq 0 ]]
}

@test "create swarm node" {
    run machine create -d $DRIVER --swarm --swarm-discovery "token://$TOKEN" workerbee
    [[ "$status" -eq 0 ]]
}

@test "ensure strategy is correct" {
    strategy=$(docker $(machine config --swarm queenbee) info | grep "Strategy:" | awk '{ print $2 }')
    echo ${strategy}
    [[ "$strategy" == "binpack" ]]
}

@test "ensure heartbeat" {
    heartbeat_arg=$(docker $(machine config queenbee) inspect -f '{{index .Args}}' swarm-agent-master)
    echo ${heartbeat_arg}
    [[ "$heartbeat_arg" =~ "--heartbeat=5s" ]]
}

@test "should not show as swarm active if normal active" {
    eval $(machine env queenbee)
    run machine ls
    echo ${output}
    [[ ${lines[1]} != *"* (swarm)"*  ]]
}

@test "should show as swarm active" {
    eval $(machine env --swarm queenbee)
    run machine ls
    echo ${output}
    [[ ${lines[1]} == *"* (swarm)"*  ]]
}

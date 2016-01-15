#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "no --help flag or command specified" {
  [[ $(machine create -d ${DRIVER} 2>&1 | grep $DRIVER | wc -l) -gt 0 ]]
}

@test "-h flag specified" {
  [[ $(machine create -d ${DRIVER} 2>&1 -h | grep $DRIVER | wc -l) -gt 0 ]]
}

@test "--help flag specified" {
  [[ $(machine create -d ${DRIVER} --help 2>&1 | grep $DRIVER | wc -l) -gt 0 ]]
}

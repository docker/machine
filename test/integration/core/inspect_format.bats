#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "$DRIVER: create" {
  run machine create -d $DRIVER $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: inspect format template" {
  run machine inspect -f '{{.DriverName}}' $NAME
  [[ "$output" == "$DRIVER" ]]
}

@test "$DRIVER: inspect format template json directive" {
  run machine inspect -f '{{json .DriverName}}' $NAME
  [[ "$output" == "\"$DRIVER\"" ]]
}

@test "$DRIVER: inspect format template pretty json directive" {
  linecount=$(machine inspect -f '{{prettyjson .Driver}}' $NAME | wc -l)
  [[ "$linecount" -gt 1 ]]
}

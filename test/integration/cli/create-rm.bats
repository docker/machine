#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "bogus: non-existent driver fails 'machine create -d bogus bogus'" {
  run machine create -d bogus bogus
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Driver \"bogus\" not found. Do you have the plugin binary accessible in your PATH?" ]]
}

@test "test: create with no name fails 'machine create -d test'" {
  run machine create -d test
  last=$((${#lines[@]} - 1))
  [ "$status" -eq 1 ]
  [[ ${lines[$last]} == "Error: No machine name specified" ]]
}

@test "test: create with invalid name fails 'machine create -d test --url none ∞'" {
  run machine create -d test --url none ∞
  last=$((${#lines[@]} - 1))
  [ "$status" -eq 1 ]
  [[ ${lines[$last]} == "Error creating machine: Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -" ]]
}

@test "test: create with invalid name fails 'machine create -d test --url none -'" {
  run machine create -d test --url none -
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Error creating machine: Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -" ]]
}

@test "test: create with invalid name fails 'machine create -d test --url none .'" {
  run machine create -d test --url none .
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Error creating machine: Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -" ]]
}

@test "test: create with invalid name fails 'machine create -d test --url none ..'" {
  run machine create -d test --url none ..
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Error creating machine: Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -" ]]
}

@test "test: create with weird but valid name succeeds 'machine create -d test --url none a'" {
  run machine create -d test --url none a
  [ "$status" -eq 0 ]
}

@test "test: name is case insensitive 'machine create -d test --url none A'" {
  skip
  run machine create -d test --url none A
  [ "$status" -eq 1 ]

  [[ ${lines[0]} == "Error creating machine: Machine A already exists" ]]
}

@test "test: fail with extra argument 'machine create -d test --url none a extra'" {
  run machine create -d test --url none a extra
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Invalid command line. Found extra arguments [extra]" ]]
}

@test "test: create with weird but valid name succeeds 'machine create -d test --url none 0'" {
  run machine create -d test --url none 0
  [ "$status" -eq 0 ]
}

@test "test: rm with no name fails 'machine rm'" {
  run machine rm
  last=$(expr ${#lines[@]} - 1)
  [ "$status" -eq 1 ]
  [[ ${lines[$last]} == "Error: Expected to get one or more machine names as arguments" ]]
}

@test "test: rm non existent machine fails 'machine rm ∞'" {
  run machine rm ∞
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Error removing host \"∞\": Host does not exist: \"∞\"" ]]
}

@test "test: rm is successful 'machine rm 0'" {
  run machine rm 0
  [ "$status" -eq 0 ]
}

# Should be replaced by the test below
@test "test: rm is successful 'machine rm a'" {
  run machine rm a
  [ "$status" -eq 0 ]
}

@test "test: rm is case insensitive 'machine rm A'" {
  skip
  run machine rm A
  [ "$status" -eq 0 ]
}

#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "bogus: non-existent driver fails 'machine create -d bogus bogus'" {
  run machine create -d bogus bogus
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Driver \"bogus\" not found. Do you have the plugin binary accessible in your PATH?" ]]
}

@test "none: create with no name fails 'machine create -d none'" {
  run machine create -d none
  last=$((${#lines[@]} - 1))
  [ "$status" -eq 1 ]
  [[ ${lines[$last]} == "Error: No machine name specified" ]]
}

@test "none: create with invalid name fails 'machine create -d none --url none ∞'" {
  run machine create -d none --url none ∞
  last=$((${#lines[@]} - 1))
  [ "$status" -eq 1 ]
  [[ ${lines[$last]} == "Error creating machine: Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -" ]]
}

@test "none: create with invalid name fails 'machine create -d none --url none -'" {
  run machine create -d none --url none -
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Error creating machine: Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -" ]]
}

@test "none: create with invalid name fails 'machine create -d none --url none .'" {
  run machine create -d none --url none .
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Error creating machine: Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -" ]]
}

@test "none: create with invalid name fails 'machine create -d none --url none ..'" {
  run machine create -d none --url none ..
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Error creating machine: Invalid hostname specified. Allowed hostname chars are: 0-9a-zA-Z . -" ]]
}

@test "none: create with weird but valid name succeeds 'machine create -d none --url none a'" {
  run machine create -d none --url none a
  [ "$status" -eq 0 ]
}

@test "none: name is case insensitive 'machine create -d none --url none A'" {
  skip
  run machine create -d none --url none A
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Error creating machine: Machine A already exists" ]]
}

@test "none: fail with extra argument 'machine create -d none --url none a extra'" {
  run machine create -d none --url none a extra
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Invalid command line. Found extra arguments [extra]" ]]
}

@test "none: create with weird but valid name succeeds 'machine create -d none --url none 0'" {
  run machine create -d none --url none 0
  [ "$status" -eq 0 ]
}

@test "none: rm with no name fails 'machine rm'" {
  run machine rm -y
  last=$(expr ${#lines[@]} - 1)
  [ "$status" -eq 1 ]
  [[ ${lines[$last]} == "Error: Expected to get one or more machine names as arguments" ]]
}

@test "none: rm non existent machine fails 'machine rm ∞'" {
  run machine rm ∞ -y
  [ "$status" -eq 1 ]
  [[ ${lines[1]} == "Error removing host \"∞\": Host does not exist: \"∞\"" ]]
}

@test "none: rm is successful 'machine rm 0'" {
  run machine rm 0 -y
  [ "$status" -eq 0 ]
}

@test "none: rm ask user confirmation when -y is not provided 'echo y | machine rm ba'" {
  run machine create -d none --url none ba
  [ "$status" -eq 0 ]
  run bash -c "echo y | machine rm ba"
  [ "$status" -eq 0 ]
}

@test "none: rm deny user confirmation when -y is not provided 'echo n | machine rm ab'" {
  run machine create -d none --url none ab
  [ "$status" -eq 0 ]
  run bash -c "echo n | machine rm ab"
  [ "$status" -eq 0 ]
}

@test "none: rm never prompt user confirmation with -f is provided 'echo n | machine rm -f ab'" {
  run machine create -d none --url none c
  [ "$status" -eq 0 ]
  run bash -c "machine rm -f c"
  [ "$status" -eq 0 ]
  [[ ${lines[1]} == "Successfully removed c" ]]
}

# Should be replaced by the test below
@test "none: rm is successful 'machine rm a'" {
  run machine rm a -y
  [ "$status" -eq 0 ]
}

@test "none: rm is case insensitive 'machine rm A'" {
  skip
  run machine rm A -y
  [ "$status" -eq 0 ]
}

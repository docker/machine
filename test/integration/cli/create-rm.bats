#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "bogus: non-existent driver fails 'machine create -d bogus bogus'" {
  run machine create -d bogus bogus
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Driver bogus not found" ]]
}

@test "none: create with no name fails 'machine create -d none " "'" {
  run machine create -d none
  last=$((${#lines[@]} - 1))
  [ "$status" -eq 1 ]
  [[ ${lines[$last]} == "You must specify a machine name" ]]
}

@test "none: create with invalid name fails 'machine create -d none --url none ∞'" {
  run machine create -d none --url none ∞
  last=$((${#lines[@]} - 1))
  [ "$status" -eq 1 ]
  [[ ${lines[$last]} == "Error creating machine: Invalid hostname specified" ]]
}

@test "none: create with invalid name fails 'machine create -d none --url none -'" {
  run machine create -d none --url none -
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Error creating machine: Invalid hostname specified" ]]
}

@test "none: create with invalid name fails 'machine create -d none --url none .'" {
  run machine create -d none --url none .
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Error creating machine: Invalid hostname specified" ]]
}

@test "none: create with invalid name fails 'machine create -d none --url none ..'" {
  run machine create -d none --url none ..
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Error creating machine: Invalid hostname specified" ]]
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

@test "none: extraneous argument is ignored in name 'machine create -d none --url none a foo'" {
  run machine create -d none a foo
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Host already exists: \"a\"" ]]
}

@test "none: create with weird but valid name succeeds 'machine create -d none --url none 0'" {
  run machine create -d none --url none 0
  [ "$status" -eq 0 ]
}

@test "none: rm with no name fails 'machine rm'" {
  run machine rm
  last=$(expr ${#lines[@]} - 1)
  [ "$status" -eq 1 ]
  [[ ${lines[$last]} == "You must specify a machine name" ]]
}

@test "none: rm non existent machine fails 'machine rm ∞'" {
  run machine rm ∞
  [ "$status" -eq 1 ]
  [[ ${lines[0]} == "Error removing machine ∞: Host does not exist: \"∞\"" ]]
}

@test "none: rm is succesful 'machine rm 0'" {
  run machine rm 0
  [ "$status" -eq 0 ]
}

# Should be replaced by the test below
@test "none: rm is succesful 'machine rm a'" {
  run machine rm a
  [ "$status" -eq 0 ]
}

@test "none: rm is case insensitive 'machine rm A'" {
  skip
  run machine rm A
  [ "$status" -eq 0 ]
}

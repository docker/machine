#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

@test "$DRIVER: create with supported engine options" {
  run machine create -d $DRIVER \
    --engine-label spam=eggs \
    --engine-storage-driver devicemapper \
    --engine-insecure-registry registry.myco.com \
    $NAME
  echo "$output"
  [ $status -eq 0 ]
}

@test "$DRIVER: check for engine label" {
  spamlabel=$(docker $(machine config $NAME) info | grep spam)
  [[ $spamlabel =~ "spam=eggs" ]]
}

@test "$DRIVER: check for engine storage driver" {
  storage_driver_info=$(docker $(machine config $NAME) info | grep "Storage Driver")
  [[ $storage_driver_info =~ "devicemapper" ]]
}

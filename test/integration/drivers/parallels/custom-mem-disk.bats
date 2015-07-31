#!/usr/bin/env bats

load ${BASE_TEST_DIR}/helpers.bash

force_env DRIVER parallels

# Default memsize is 1024MB and disksize is 20000MB
# These values are defined in drivers/parallels/parallels.go
export DEFAULT_MEMSIZE=1024
export DEFAULT_DISKSIZE=20000
export CUSTOM_MEMSIZE=1536
export CUSTOM_DISKSIZE=10000
export CUSTOM_CPUCOUNT=1

function findDiskSize() {
  run bash -c "prlctl list -i $NAME | grep 'hdd0.*sata' | grep -o '\d*Mb' | awk -F 'Mb' '{print $1}'"
}

function findMemorySize() {
  run bash -c "prlctl list -i $NAME | grep 'memory ' | grep -o '[0-9]\+'"
}

function findCPUCount() {
  run bash -c "prlctl list -i $NAME | grep -o 'cpus=\d*' | cut -d'=' -f2"
}

@test "$DRIVER: create with custom disk, cpu count and memory size flags" {
  run machine create -d $DRIVER --parallels-cpu-count $CUSTOM_CPUCOUNT --parallels-disk-size $CUSTOM_DISKSIZE --parallels-memory $CUSTOM_MEMSIZE $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: check custom machine memory size" {
  findMemorySize
  [[ ${output} == "$CUSTOM_MEMSIZE"  ]]
}

@test "$DRIVER: check custom machine disksize" {
  findDiskSize
  [[ ${output} == *"$CUSTOM_DISKSIZE"* ]]
}

@test "$DRIVER: check custom machine cpucount" {
  findCPUCount
  [[ ${output} == "$CUSTOM_CPUCOUNT" ]]
}

@test "$DRIVER: machine should show running after create" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

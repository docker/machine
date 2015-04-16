#!/usr/bin/env bats

load helpers

export DRIVER=virtualbox
export NAME="bats-$DRIVER-test"
export MACHINE_STORAGE_PATH=/tmp/machine-bats-test-$DRIVER
# Default memsize is 1024MB and disksize is 20000MB
# These values are defined in drivers/virtualbox/virtualbox.go
export DEFAULT_MEMSIZE=1024
export DEFAULT_DISKSIZE=20000
export CUSTOM_MEMSIZE=1536
export CUSTOM_DISKSIZE=10000
export CUSTOM_CPUCOUNT=1
export BAD_URL="http://dev.null:9111/bad.iso"

function setup() {
  # add sleep because vbox; ugh
  sleep 1
}

findDiskSize() {
  # SATA-0-0 is usually the boot2disk.iso image
  # We assume that SATA 1-0 is root disk VMDK and grab this UUID
  # e.g. "SATA-ImageUUID-1-0"="fb5f33a7-e4e3-4cb9-877c-f9415ae2adea"
  # TODO(slashk): does this work on Windows ?
  run bash -c "VBoxManage showvminfo --machinereadable $NAME | grep SATA-ImageUUID-1-0 | cut -d'=' -f2"
  run bash -c "VBoxManage showhdinfo $output | grep "Capacity:" | awk -F' ' '{ print $2 }'"
}

findMemorySize() {
  run bash -c "VBoxManage showvminfo --machinereadable $NAME | grep memory= | cut -d'=' -f2"
}

findCPUCount() {
  run bash -c "VBoxManage showvminfo --machinereadable $NAME | grep cpus= | cut -d'=' -f2"
}

buildMachineWithOldIsoCheckUpgrade() {
  run wget https://github.com/boot2docker/boot2docker/releases/download/v1.4.1/boot2docker.iso -O $MACHINE_STORAGE_PATH/cache/boot2docker.iso
  run machine create -d virtualbox $NAME
  run machine upgrade $NAME
}

@test "$DRIVER: VM should not exist" {
  run VBoxManage showvminfo $NAME
  [ "$status" -eq 1  ]
}

load common-create

@test "$DRIVER: machine should not exist" {
  run machine active $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: check default machine memory size" {
  findMemorySize
  [[ ${output} == "${DEFAULT_MEMSIZE}"  ]]
}

@test "$DRIVER: check default machine disksize" {
  findDiskSize
  [[ ${output} == *"$DEFAULT_DISKSIZE"* ]]
}

load common-start-stop

@test "$DRIVER: VBoxManage pause" {
  run VBoxManage controlvm $NAME pause
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show paused after VBoxManage pause" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Paused"*  ]]
}

@test "$DRIVER: start after paused" {
  run machine start $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show running after start" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

@test "$DRIVER: VBoxManage savestate" {
  run VBoxManage controlvm $NAME savestate
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show saved after VBoxManage savestate" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"$NAME"*  ]]
  [[ ${lines[1]} == *"Saved"*  ]]
}

@test "$DRIVER: start after saved" {
  run machine start $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should show running after start" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

load common-destroy

@test "$DRIVER: VM should not exist after remove" {
  run VBoxManage showvminfo $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: create too small disk size" {
  run machine create -d $DRIVER --virtualbox-disk-size 0 $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: remove after too small create" {
  run machine rm -f $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: create too large disk size" {
  skip "this will take too long to run effectively"
  run machine create -d $DRIVER --virtualbox-disk-size 1000000 $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: remove after too large create" {
  skip "no need to remove if large test not run"
  run machine rm -f $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: should not create with incorrect value type for disk size" {
  run machine create -d $DRIVER --virtualbox-disk-size ffsfwf $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: create too small memory size" {
  run machine create -d $DRIVER --virtualbox-memory 0 $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: remove after too small memory" {
  run machine rm -f $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: create with bad boot2docker url" {
  run machine create -d $DRIVER --virtualbox-boot2docker-url $BAD_URL $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: remove after bad boot2docker url" {
  run machine rm -f $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: create with custom disk, cpu count and memory size flags" {
  run machine create -d $DRIVER --virtualbox-cpu-count $CUSTOM_CPUCOUNT --virtualbox-disk-size $CUSTOM_DISKSIZE --virtualbox-memory $CUSTOM_MEMSIZE $NAME
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

@test "$DRIVER: remove after custom flag create" {
  run machine rm -f $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: can create custom machine using disk size, cpu count and memory size via env vars" {
  VIRTUALBOX_DISK_SIZE=$CUSTOM_DISKSIZE VIRTUALBOX_CPU_COUNT=$CUSTOM_CPUCOUNT VIRTUALBOX_MEMORY_SIZE=$CUSTOM_MEMSIZE run machine create -d $DRIVER $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: check machine's memory size was set correctly by env var" {
  findMemorySize
  [[ ${output} == "$CUSTOM_MEMSIZE"  ]]
}

@test "$DRIVER: check machine's disk size was set correctly by env var" {
  findDiskSize
  [[ ${output} == *"$CUSTOM_DISKSIZE"* ]]
}

@test "$DRIVER: check custom machine cpucount" {
  findCPUCount
  [[ ${output} == "$CUSTOM_CPUCOUNT" ]]
}


@test "$DRIVER: machine should show running after create with env" {
  run machine ls
  [ "$status" -eq 0  ]
  [[ ${lines[1]} == *"Running"*  ]]
}

@test "$DRIVER: remove after custom env create" {
  run machine rm -f $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: upgrade should work" {
  buildMachineWithOldIsoCheckUpgrade
  [ "$status" -eq 0 ]
}

@test "$DRIVER: remove machine after upgrade test" {
  run machine rm -f $NAME
}

# Cleanup of machine store should always be the last 'test'
@test "$DRIVER: cleanup" {
  run rm -rf $MACHINE_STORAGE_PATH
  [ "$status" -eq 0  ]
}


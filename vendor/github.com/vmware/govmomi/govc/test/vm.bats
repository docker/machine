#!/usr/bin/env bats

load test_helper

@test "vm.ip" {
  id=$(new_ttylinux_vm)

  run govc vm.power -on $id
  assert_success

  run govc vm.ip $id
  assert_success
}

@test "vm.ip -esxcli" {
  id=$(new_ttylinux_vm)

  run govc vm.power -on $id
  assert_success

  run govc vm.ip -esxcli $id
  assert_success

  ip_esxcli=$output

  run govc vm.ip $id
  assert_success
  ip_tools=$output

  assert_equal $ip_esxcli $ip_tools
}

@test "vm.create" {
  id=$(new_ttylinux_vm)

  run govc vm.power -on $id
  assert_success

  result=$(govc device.ls -vm $vm | grep disk- | wc -l)
  [ $result -eq 0 ]

  result=$(govc device.ls -vm $vm | grep cdrom- | wc -l)
  [ $result -eq 0 ]
}

@test "vm.change" {
  id=$(new_ttylinux_vm)

  run govc vm.change -g ubuntu64Guest -m 1024 -c 2 -vm $id
  assert_success

  run govc vm.info $id
  assert_success
  assert_line "Guest name: Ubuntu Linux (64-bit)"
  assert_line "Memory: 1024MB"
  assert_line "CPU: 2 vCPU(s)"

  run govc vm.change -e "guestinfo.a=1" -e "guestinfo.b=2" -vm $id
  assert_success

  run govc vm.info -e $id
  assert_success
  assert_line "guestinfo.a: 1"
  assert_line "guestinfo.b: 2"

  nid=$(new_id)
  run govc vm.change -name $nid -vm $id
  assert_success

  run govc vm.info $id
  [ ${#lines[@]} -eq 0 ]

  run govc vm.info $nid
  [ ${#lines[@]} -gt 0 ]
}

@test "vm.power" {
  vm=$(new_ttylinux_vm)

  run vm_power_state $vm
  assert_success "poweredOff"

  run govc vm.power $vm
  assert_failure

  run govc vm.power -on -off $vm
  assert_failure

  run govc vm.power -on $vm
  assert_success
  run vm_power_state $vm
  assert_success "poweredOn"

  run govc vm.power -suspend $vm
  assert_success
  run vm_power_state $vm
  assert_success "suspended"

  run govc vm.power -on $vm
  assert_success
  run vm_power_state $vm
  assert_success "poweredOn"
}

@test "vm.create pvscsi" {
  vm=$(new_id)
  govc vm.create -on=false -disk.controller pvscsi $vm

  result=$(govc device.ls -vm $vm | grep pvscsi- | wc -l)
  [ $result -eq 1 ]

  result=$(govc device.ls -vm $vm | grep lsilogic- | wc -l)
  [ $result -eq 0 ]
}

@test "vm.create in cluster" {
  vcsim_env

  # using GOVC_HOST and its resource pool
  run govc vm.create -on=false $(new_id)
  assert_success

  # using no -host and the default resource pool for DC0
  unset GOVC_HOST
  run govc vm.create -on=false $(new_id)
  assert_success
}

@test "vm.info" {
  local num=3

  local prefix=$(new_id)

  for x in $(seq $num)
  do
    local id="${prefix}-${x}"
    run govc vm.create -on=false $id
    assert_success

    local found=$(govc vm.info $id | grep Name: | wc -l)
    [ "$found" -eq 1 ]
  done

  # test find slice
  local found=$(govc vm.info ${prefix}-* | grep Name: | wc -l)
  [ "$found" -eq $num ]

  # test extraConfig
  run govc vm.change -e "guestinfo.a=2" -vm $id
  assert_success
  run govc vm.info -e $id
  assert_success
  assert_line "guestinfo.a: 2"
}

@test "vm.create linked ide disk" {
  vm=$(new_id)
  run govc vm.create -disk $GOVC_TEST_VMDK -disk.controller ide -on=false $vm
  assert_success

  run govc device.info -vm $vm disk-200-0
  assert_success
  assert_line "Controller: ide-200"
}

@test "vm.create linked scsi disk" {
  vm=$(new_id)

  run govc vm.create -disk enoent -on=false $vm
  assert_failure "Error: datastore file does not exist"

  run govc vm.create -disk $GOVC_TEST_VMDK -on=false $vm
  assert_success

  run govc device.info -vm $vm disk-1000-0
  assert_success
  assert_line "Controller: lsilogic-1000"
  assert_line "Parent: [datastore1] $GOVC_TEST_VMDK"
  assert_line "File: [datastore1] $vm/${vm}.vmdk"
}

@test "vm.create scsi disk" {
  vm=$(new_id)

  run govc vm.create -disk enoent -on=false $vm
  assert_failure "Error: datastore file does not exist"

  run govc vm.create -disk $GOVC_TEST_VMDK -on=false -link=false $vm
  assert_success

  run govc device.info -vm $vm disk-1000-0
  assert_success
  assert_line "Controller: lsilogic-1000"
  refute_line "Parent: [datastore1] $GOVC_TEST_VMDK"
  assert_line "File: [datastore1] $GOVC_TEST_VMDK"
}

@test "vm.create iso" {
  upload_iso

  vm=$(new_id)

  run govc vm.create -iso enoent -on=false $vm
  assert_failure "Error: datastore file does not exist"

  run govc vm.create -iso $GOVC_TEST_ISO -on=false $vm
  assert_success

  run govc device.info -vm $vm cdrom-3000
  assert_success
  assert_line "Controller: ide-200"
  assert_line "Summary: ISO [datastore1] $GOVC_TEST_ISO"
}

@test "vm.disk.create empty vm" {
  vm=$(new_empty_vm)

  local name=$(new_id)

  run govc vm.disk.create -vm $vm -name $name -size 1G
  assert_success
  result=$(govc device.ls -vm $vm | grep disk- | wc -l)
  [ $result -eq 1 ]

  name=$(new_id)

  run govc vm.disk.create -vm $vm -name $name -controller lsilogic-1000 -size 2G
  assert_success
  result=$(govc device.ls -vm $vm | grep disk- | wc -l)
  [ $result -eq 2 ]
}

@test "vm.disk.create" {
  import_ttylinux_vmdk
  vm=$(new_id)

  govc vm.create -disk $GOVC_TEST_VMDK -on=false $vm
  result=$(govc device.ls -vm $vm | grep disk- | wc -l)
  [ $result -eq 1 ]

  local name=$(new_id)

  run govc vm.disk.create -vm $vm -name $name -size 1G
  assert_success
  result=$(govc device.ls -vm $vm | grep disk- | wc -l)
  [ $result -eq 2 ]

  run govc vm.disk.create -vm $vm -name $name -size 1G
  assert_success # TODO: should fail?
  result=$(govc device.ls -vm $vm | grep disk- | wc -l)
  [ $result -eq 2 ]
}

@test "vm.disk.attach" {
  import_ttylinux_vmdk
  vm=$(new_id)

  govc vm.create -disk $GOVC_TEST_VMDK -on=false $vm
  result=$(govc device.ls -vm $vm | grep disk- | wc -l)
  [ $result -eq 1 ]

  run govc import.vmdk $GOVC_TEST_VMDK_SRC $vm
  assert_success

  run govc vm.disk.attach -vm $vm -link=false -disk enoent.vmdk
  assert_failure "Error: File [datastore1] enoent.vmdk was not found"

  run govc vm.disk.attach -vm $vm -disk enoent.vmdk
  assert_failure "Error: Invalid configuration for device '0'."

  run govc vm.disk.attach -vm $vm -disk $vm/$GOVC_TEST_VMDK -controller lsilogic-1000
  assert_success
  result=$(govc device.ls -vm $vm | grep disk- | wc -l)
  [ $result -eq 2 ]
}

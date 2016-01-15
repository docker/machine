#!/usr/bin/env bats

load test_helper

@test "import.ova" {
  run govc import.ova $GOVC_IMAGES/${TTYLINUX_NAME}.ova
  assert_success

  run govc vm.destroy ${TTYLINUX_NAME}
  assert_success
}

@test "import.ova with iso" {
  run govc import.ova $GOVC_IMAGES/${TTYLINUX_NAME}-live.ova
  assert_success

  run govc vm.destroy ${TTYLINUX_NAME}-live
  assert_success
}

@test "import.ovf" {
  run govc import.ovf $GOVC_IMAGES/${TTYLINUX_NAME}.ovf
  assert_success

  run govc vm.destroy ${TTYLINUX_NAME}
  assert_success

  # test w/ relative dir
  pushd $BATS_TEST_DIRNAME >/dev/null
  run govc import.ovf ./images/${TTYLINUX_NAME}.ovf
  assert_success
  popd >/dev/null

  run govc vm.destroy ${TTYLINUX_NAME}
  assert_success
}

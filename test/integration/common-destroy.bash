@test "$DRIVER: remove" {
  run machine rm -f $NAME
  [ "$status" -eq 0  ]
}

@test "$DRIVER: machine should not exist" {
  run machine active $NAME
  [ "$status" -eq 1  ]
}

@test "$DRIVER: cleanup" {
  run rm -rf $MACHINE_STORAGE_PATH
  [ "$status" -eq 0  ]
}

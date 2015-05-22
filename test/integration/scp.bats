#!/usr/bin/env bats

load helpers

export DRIVER=virtualbox
export NAME="bats-$DRIVER-test"
export MACHINE_STORAGE_PATH=/tmp/machine-bats-test-$DRIVER
export SECOND_MACHINE="$NAME-2"

@test "$DRIVER: create" {
  run machine create -d $DRIVER $NAME
  [[ ${status} -eq 0 ]]
}

@test "$DRIVER: test machine scp command from remote to host" {
  machine ssh $NAME 'echo A file created remotely! >/tmp/foo.txt'
  machine scp $NAME:/tmp/foo.txt .
  [[ $(cat foo.txt) == "A file created remotely!" ]]
}

@test "$DRIVER: test machine scp command from host to remote" {
  echo A file created locally! >foo.txt
  machine scp foo.txt $NAME:/tmp/foo.txt
  [[ $(machine ssh $NAME cat /tmp/foo.txt) == "A file created locally!" ]]
}

@test "$DRIVER: create machine to test transferring files from machine to machine" {
  run machine create -d $DRIVER $SECOND_MACHINE
  [[ ${status} -eq 0 ]]
}

@test "$DRIVER: scp from one machine to another" {
  run machine ssh $NAME 'echo A file hopping around! >/tmp/foo.txt'
  run machine scp $NAME:/tmp/foo.txt $SECOND_MACHINE:/tmp/foo.txt
  [[ $(machine ssh ${SECOND_MACHINE} cat /tmp/foo.txt) == "A file hopping around!" ]]
}

@test "cleanup" {
  rm foo.txt
  machine rm $NAME
}

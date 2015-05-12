#!/usr/bin/env bats

load helpers

export DRIVER=virtualbox
export NAME="bats-$DRIVER-test"
export MACHINE_STORAGE_PATH=/tmp/machine-bats-test-$DRIVER

@test "$DRIVER: create" {
  run machine create -d $DRIVER $NAME
}

@test "$DRIVER: verify that server cert checksum matches local checksum" {
  # TODO: Does this test work OK on Linux? cc @ehazlett
  # Have to create this directory and file or else the OpenSSL checksum will barf.
  machine ssh $NAME -- sudo mkdir -p /usr/local/ssl
  machine ssh $NAME -- sudo touch /usr/local/ssl/openssl.cnf
  SERVER_CHECKSUM=$(machine ssh $NAME -- openssl dgst -sha256 /var/lib/boot2docker/ca.pem | awk '{ print $2 }')
  LOCAL_CHECKSUM=$(openssl dgst -sha256 $MACHINE_STORAGE_PATH/certs/ca.pem | awk '{ print $2 }')
  echo ${SERVER_CHECKSUM}
  echo ${LOCAL_CHECKSUM}
  [[ ${SERVER_CHECKSUM} == ${LOCAL_CHECKSUM} ]]
}

@test "cleanup" {
  machine rm $NAME
}

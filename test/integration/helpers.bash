#!/bin/bash

# Root directory of the repository.
MACHINE_ROOT=${BATS_TEST_DIRNAME}/../..

PLATFORM=`uname -s | tr '[:upper:]' '[:lower:]'`
ARCH=`uname -m`

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
else
    ARCH="386"
fi
MACHINE_BIN_NAME=docker-machine_$PLATFORM-$ARCH

build_machine() {
    pushd $MACHINE_ROOT >/dev/null
    godep go build -o $MACHINE_BIN_NAME
    popd >/dev/null
}

# build machine binary if needed
if [ ! -e $MACHINE_ROOT/$MACHINE_BIN_NAME ]; then
    build_machine
fi

function machine() {
    ${MACHINE_ROOT}/$MACHINE_BIN_NAME "$@"
}

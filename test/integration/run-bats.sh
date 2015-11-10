#!/bin/bash

set -e

# Wrapper script to run bats tests for various drivers.
# Usage: DRIVER=[driver] ./run-bats.sh [subtest]
# or
# If you just want to quickly run the test on an existing container
# Usage: RECYCLE_MACHINE=[container name] ./run-bats.sh [subtest]

function quiet_run () {
    if [[ "$VERBOSE" == "1" ]]; then
        "$@"
    else
        "$@" &>/dev/null
    fi
}

function cleanup_machines() {
    if [[ -z "$RECYCLE_MACHINE" ]]; then
        if [[ $(machine ls -q | wc -l) -ne 0 ]]; then
            quiet_run machine rm -f $(machine ls -q)
        fi
    else
        echo "Explicity using Container ${RECYCLE_MACHINE}, so not killing it, on your request..."
    fi
}

function machine() {
    export PATH="$MACHINE_ROOT"/bin:$PATH
    "$MACHINE_ROOT"/bin/"$MACHINE_BIN_NAME" "$@"
}

function run_bats() {
    for bats_file in $(find "$1" -name \*.bats); do
        # BATS returns non-zero to indicate the tests have failed, we shouldn't
        # neccesarily bail in this case, so that's the reason for the e toggle.
        set +e
        echo "=> $bats_file"
        bats "$bats_file"
        if [[ $? -ne 0 ]]; then
            EXIT_STATUS=1
        fi
        set -e
        echo
        cleanup_machines
    done
}

# Set this ourselves in case bats call fails
EXIT_STATUS=0
export BATS_FILE="$1"

if [[ -z "$RECYCLE_MACHINE" ]] && [[ -z "$DRIVER" ]]; then
    echo "You must specify the DRIVER environment variable."
    exit 1
fi

if [[ -z "$BATS_FILE" ]]; then
    echo "You must specify a bats test to run."
    exit 1
fi

if [[ ! -e "$BATS_FILE" ]]; then
    echo "Requested bats file or directory not found: $BATS_FILE"
    exit 1
fi

# TODO: Should the script bail out if these are set already?
export BASE_TEST_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
export MACHINE_ROOT="$BASE_TEST_DIR/../.."

if [[ -z "$RECYCLE_MACHINE" ]]; then
    export MACHINE_STORAGE_PATH="/tmp/machine-bats-test-$DRIVER"
    export NAME="bats-$DRIVER-test"
else
    export DRIVER=$(docker-machine inspect "$RECYCLE_MACHINE" | grep DriverName | cut -d\" -f 4)
    export NAME=$RECYCLE_MACHINE
fi

export MACHINE_BIN_NAME=docker-machine
export BATS_LOG="$MACHINE_ROOT/bats.log"
B2D_LOCATION=~/.docker/machine/cache/boot2docker.iso

# This function gets used in the integration tests, so export it.
export -f machine

touch "$BATS_LOG"
rm "$BATS_LOG"

if [[ -z "$RECYCLE_MACHINE" ]]; then

    if [[ -d "$MACHINE_STORAGE_PATH" ]]; then
        rm -r "$MACHINE_STORAGE_PATH"
    fi

    if [[ "$b2dcache" == "1" ]] && [[ -f $B2D_LOCATION ]]; then
        mkdir -p "${MACHINE_STORAGE_PATH}/cache"
        cp $B2D_LOCATION "${MACHINE_STORAGE_PATH}/cache/boot2docker.iso"
    fi

fi

run_bats "$BATS_FILE"

if [[ -z "$RECYCLE_MACHINE" ]]; then

    if [[ -d "$MACHINE_STORAGE_PATH" ]]; then
        rm -r "$MACHINE_STORAGE_PATH"
    fi
fi

set +e
pkill docker-machine
if [[ $? -eq 0 ]]; then
    EXIT_STATUS=1
fi
set -e

exit ${EXIT_STATUS}

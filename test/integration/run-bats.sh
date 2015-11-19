#!/bin/bash

set -e

# Wrapper script to run bats tests for various drivers.
# Usage: DRIVER=[driver] ./run-bats.sh [subtest]

function quiet_run () {
    if [[ "$VERBOSE" == "1" ]]; then
        "$@"
    else
        "$@" &>/dev/null
    fi
}

function cleanup_machines() {
    if [[ $(machine ls -q | wc -l) -ne 0 ]]; then
        quiet_run machine rm -f $(machine ls -q)
    fi
}

function cleanup_store() {
    if [[ -d "$MACHINE_STORAGE_PATH" ]]; then
        rm -r "$MACHINE_STORAGE_PATH"
    fi
}

function machine() {
    export PATH="$MACHINE_ROOT"/bin:$PATH
    "$MACHINE_ROOT"/bin/"$MACHINE_BIN_NAME" "$@"
}

function run_bats() {
    for bats_file in $(find "$1" -name \*.bats); do
        export NAME="bats-$DRIVER-test-$(date +%s)"

        # BATS returns non-zero to indicate the tests have failed, we shouldn't
        # neccesarily bail in this case, so that's the reason for the e toggle.
        echo "=> $bats_file"

        set +e
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

if [[ -z "$DRIVER" ]]; then
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
export MACHINE_STORAGE_PATH="/tmp/machine-bats-test-$DRIVER"
export MACHINE_BIN_NAME=docker-machine
export BATS_LOG="$MACHINE_ROOT/bats.log"
export B2D_LOCATION=~/.docker/machine/cache/boot2docker.iso

# This function gets used in the integration tests, so export it.
export -f machine

> "$BATS_LOG"

cleanup_machines
cleanup_store

if [[ "$B2D_CACHE" == "1" ]] && [[ -f $B2D_LOCATION ]]; then
    mkdir -p "${MACHINE_STORAGE_PATH}/cache"
    cp $B2D_LOCATION "${MACHINE_STORAGE_PATH}/cache/boot2docker.iso"
fi

run_bats "$BATS_FILE"

cleanup_store

exit ${EXIT_STATUS}

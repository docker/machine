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
        quiet_run machine stop $(machine ls -q)
        quiet_run machine rm $(machine ls -q)
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
export NAME="bats-$DRIVER-test"
export MACHINE_STORAGE_PATH="/tmp/machine-bats-test-$DRIVER"
export MACHINE_BIN_NAME=docker-machine
export BATS_LOG="$MACHINE_ROOT/bats.log"

# This function gets used in the integration tests, so export it.
export -f machine

touch "$BATS_LOG"
rm "$BATS_LOG"

run_bats "$BATS_FILE"

if [[ -d "$MACHINE_STORAGE_PATH" ]]; then
    rm -r "$MACHINE_STORAGE_PATH"
fi

set +e
pkill docker-machine
if [[ $? -eq 0 ]]; then
    EXIT_STATUS=1
fi
set -e

exit ${EXIT_STATUS}

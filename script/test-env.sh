#!/bin/sh

MACHINE_BIN=./machine_darwin_amd64

# Clean before all to ensure consistency
echo "== Cleaning before all"
$MACHINE_BIN rm -f dev

######################## virtualbox driver tests

echo "== Basic test with a string provided and clean"
$MACHINE_BIN create -d virtualbox --docker-opts a dev
$MACHINE_BIN rm -f dev

echo "== Non regression Test without the args and clean"
$MACHINE_BIN create -d virtualbox dev
$MACHINE_BIN rm -f dev

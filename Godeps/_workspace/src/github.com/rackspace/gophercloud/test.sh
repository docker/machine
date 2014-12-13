#!/bin/bash

source demo-openrc.sh
clear
go test -v -tags 'acceptance fixtures' ./acceptance/rackspace/compute/v2/... -run ServerOperations

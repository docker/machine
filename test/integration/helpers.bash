#!/bin/bash

teardown() {
  echo "$BATS_TEST_NAME
----------
$output
----------

" >> ${BATS_LOG}
}

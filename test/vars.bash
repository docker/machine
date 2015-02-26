#!/bin/bash

PLATFORM=`uname -s | tr '[:upper:]' '[:lower:]'`
ARCH=`uname -m`

if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
else
    ARCH="386"
fi

# Project name, used to name the binaries
PKG_NAME := docker-machine
GH_USER ?= docker
GH_REPO ?= machine

# If true, disable optimizations and does NOT strip the binary
DEBUG ?=
# If true, "build" will produce a static binary (cross compile always produce static build regardless)
STATIC ?=
# If true, turn on verbose output for build
VERBOSE ?=
# Build tags
BUILDTAGS ?=
# Adjust parallelism for gox
PARALLEL ?= -1
# Coverage default directory
COVERAGE_DIR ?= cover
# Whether to perform targets inside a docker container, or natively on the host
USE_CONTAINER ?=

# List of cross compilation targets
ifeq ($(TARGET_OS),)
  TARGET_OS := darwin freebsd linux windows
endif

ifeq ($(TARGET_ARCH),)
  TARGET_ARCH := amd64 arm 386
endif

# Output prefix, defaults to local directory if not specified
ifeq ($(PREFIX),)
  PREFIX := $(shell pwd)
endif

default: build
clean: coverage-clean build-clean
build: build-simple
test: dco fmt vet test-short
validate: dco fmt vet lint test-short test-long
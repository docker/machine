# Project name, used to name the binaries
PKG_NAME := docker-machine
# Github infos
GITHUB_INFO := $(shell git config --get remote.origin.url | sed -e 's/.*:\(.*\).git/\1/')
GH_USER ?= $(shell echo $(GITHUB_INFO) | cut -d \/ -f 1)
GH_REPO ?= $(shell echo $(GITHUB_INFO) | cut -d \/ -f 2)

# If true, disable optimizations and does NOT strip the binary
DEBUG ?=
# If true, "build" will produce a static binary (cross compile always produce static build regardless)
STATIC ?=
# If true, turn on verbose output for build
VERBOSE ?=
# Build tags
BUILDTAGS ?=
# Adjust number of parallel builds (XXX not used)
PARALLEL ?= -1
# Coverage default directory
COVERAGE_DIR ?= cover
# Whether to perform targets inside a docker container, or natively on the host
USE_CONTAINER ?=

# List of cross compilation targets
ifeq ($(TARGET_OS),)
  TARGET_OS := darwin linux windows
endif

ifeq ($(TARGET_ARCH),)
  TARGET_ARCH := amd64 386
endif

# Output prefix, defaults to local directory if not specified
ifeq ($(PREFIX),)
  PREFIX := $(shell pwd)
endif

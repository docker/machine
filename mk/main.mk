# Initialize version and gc flags
GO_LDFLAGS := -X `go list ./version`.GitCommit=`git rev-parse --short HEAD`
GO_GCFLAGS :=

# Full package list
PKGS := $(shell go list -tags "$(BUILDTAGS)" ./commands/... ./libmachine/... ./drivers/... | grep -v "/vendor/" | grep -v "/Godeps/")

# Support go1.5 vendoring (let us avoid messing with GOPATH or using godep)
export GO15VENDOREXPERIMENT = 1

# Resolving binary dependencies for specific targets
GOLINT_BIN := $(GOPATH)/bin/golint
GOLINT := $(shell [ -x $(GOLINT_BIN) ] && echo $(GOLINT_BIN) || echo '')

# Honor debug
ifeq ($(DEBUG),true)
	# Disable function inlining and variable registerization
	GO_GCFLAGS := -gcflags "-N -l"
else
	# Turn of DWARF debugging information and strip the binary otherwise
	GO_LDFLAGS := $(GO_LDFLAGS) -w -s
endif

# Honor static
ifeq ($(STATIC),true)
	# Append to the version
	GO_LDFLAGS := $(GO_LDFLAGS) -extldflags -static
endif

# Honor verbose
VERBOSE_GO := 
GO := go
ifeq ($(VERBOSE),true)
	VERBOSE_GO := -v
	GO := go
endif

include mk/build.mk
include mk/coverage.mk
include mk/release.mk
include mk/test.mk
include mk/validate.mk

.all_build: build build-clean build-x build-machine build-plugins
.all_coverage: coverage-generate coverage-html coverage-send coverage-serve coverage-clean
.all_release: release-checksum release
.all_test: test-short test-long test-integration
.all_validate: dco fmt vet lint

default: build
# Build native machine and all drivers
build: build-machine build-plugins
# Just build native machine itself
machine: build-machine
# Just build the native plugins
plugins: build-plugins
# Build all, cross platform
cross: build-x

clean: coverage-clean build-clean
test: dco fmt vet test-short
validate: dco fmt vet lint test-short test-long
install:
	cp ./bin/docker-machine* /usr/local/bin/

.PHONY: .all_build .all_coverage .all_release .all_test .all_validate test build validate clean

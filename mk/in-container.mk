DOCKER_IMAGE_NAME := "docker-machine-build"
DOCKER_CONTAINER_NAME := "docker-machine-build-container"

.SILENT:
.PHONY: test

test clean build machine plugins cross validate dco fmt vet test-short test-long coverage-send: all

all:
	@test -z '$(shell docker ps -a | grep $(DOCKER_CONTAINER_NAME))' || docker rm -f $(DOCKER_CONTAINER_NAME)

	docker build -t $(DOCKER_IMAGE_NAME) .
	docker run --name $(DOCKER_CONTAINER_NAME) \
	    -e DEBUG \
	    -e STATIC \
	    -e VERBOSE \
	    -e BUILDTAGS \
	    -e PARALLEL \
	    -e COVERAGE_DIR \
	    -e TARGET_OS \
	    -e TARGET_ARCH \
	    -e PREFIX \
	    $(DOCKER_IMAGE_NAME) \
	    make $(MAKECMDGOALS)

	@test ! -d bin || rm -Rf bin
	@(docker cp $(DOCKER_CONTAINER_NAME):/go/src/github.com/docker/machine/bin bin >/dev/null 2>&1) || true

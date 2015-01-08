
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null)
IMAGENAME := docker-machine$(if $(GIT_BRANCH),:$(GIT_BRANCH))

all: linux darwin

build:
	docker build -t "$(IMAGENAME)" .

linux: build
	docker run --name "linuxmachine" -it  "$(IMAGENAME)" go build
	docker cp "linuxmachine":/go/src/github.com/docker/machine/machine machine-linux
	docker rm "linuxmachine"

darwin: build
	docker run --name "darwinmachine" -it -e GOOS=darwin "$(IMAGENAME)" go build
	docker cp "darwinmachine":/go/src/github.com/docker/machine/machine machine-darwin
	docker rm "darwinmachine"

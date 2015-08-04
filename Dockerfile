FROM golang:1.4.2-cross

# TODO: Vendor these `go get` commands using Godep.
RUN go get github.com/mitchellh/gox
RUN go get github.com/aktau/github-release
RUN go get github.com/tools/godep
RUN go get golang.org/x/tools/cmd/cover

ENV GOPATH /go/src/github.com/docker/machine/Godeps/_workspace:/go
ENV MACHINE_BINARY /go/src/github.com/docker/machine/docker-machine
ENV USER root

WORKDIR /go/src/github.com/docker/machine

ADD . /go/src/github.com/docker/machine

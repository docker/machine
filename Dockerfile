FROM golang:1.3-cross
RUN go get github.com/mitchellh/gox
ENV GOPATH /go/src/github.com/docker/machine/Godeps/_workspace:/go
WORKDIR /go/src/github.com/docker/machine
ADD . /go/src/github.com/docker/machine

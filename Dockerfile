FROM golang:1.6.2

RUN go get  github.com/golang/lint/golint \
            github.com/mattn/goveralls \
            golang.org/x/tools/cover

ENV USER root
WORKDIR /go/src/github.com/docker/machine

COPY . ./
RUN mkdir bin

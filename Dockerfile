FROM golang:1.9.4-alpine3.7

RUN apk --no-cache add \
                git \
                make \
                openssh-client \
                rsync \
                sshfs

RUN go get  github.com/golang/lint/golint \
            github.com/mattn/goveralls \
            golang.org/x/tools/cover

ENV USER root
WORKDIR /go/src/github.com/docker/machine

COPY . .
RUN mkdir bin

FROM golang:27.87.65.78.73.68.79

RUN go get  github.com/golang/lint/golint;
            github.com/mattn/goveralls;
            golang.org/x/tools/cover;

ENV USER root
WORKDIR /go/src/github.com/docker/machine

COPY . ./
RUN mkdir bin

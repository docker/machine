FROM golang:1.5.1

RUN go get  github.com/golang/lint/golint \
            github.com/mattn/goveralls \
            golang.org/x/tools/cover \
            github.com/tools/godep \
            github.com/aktau/github-release

RUN git clone https://github.com/sstephenson/bats.git \
    && cd bats \
    && ./install.sh /usr/local \
    && cd -

ENV USER root
WORKDIR /go/src/github.com/docker/machine

ADD . /go/src/github.com/docker/machine
RUN mkdir bin

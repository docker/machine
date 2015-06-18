.PHONY: all test validate-dco validate-gofmt validate build

all: validate test build

test:
	script/test

validate-dco:
	script/validate-dco

validate-gofmt:
	script/validate-gofmt

validate: validate-dco validate-gofmt

build:
	script/build


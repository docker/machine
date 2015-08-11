.PHONY: test validate-dco validate-gofmt

default: build

remote: build-remote

test:
	script/test

validate-dco:
	script/validate-dco

validate-gofmt:
	script/validate-gofmt

validate: validate-dco validate-gofmt test

build: clean
	script/build

build-remote: clean
	script/build-remote

clean:
	rm -f docker-machine_*
	rm -rf Godeps/_workspace/pkg

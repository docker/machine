# Validate DCO on all history
dco:
	@script/validate-dco

# Fmt
fmt:
	@test -z "$$(gofmt -s -l . 2>&1 | grep -v vendor/ | grep -v Godeps/ | tee /dev/stderr)"

# Vet
vet: build
	@test -z "$$(go vet $(PKGS) 2>&1 | tee /dev/stderr)"

# Lint
lint:
	$(if $(GOLINT), , \
		$(error Please install golint: go get -u github.com/golang/lint/golint))
	@test -z "$$($(GOLINT) ./... 2>&1 | grep -v vendor/ | grep -v Godeps/ | tee /dev/stderr)"


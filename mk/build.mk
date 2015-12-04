extension = $(patsubst windows,.exe,$(filter windows,$(1)))

define gocross
	GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 $(GO) build \
		-o $(PREFIX)/bin/docker-machine_$(1)-$(2)/docker-$(patsubst cmd/%.go,%,$3)$(call extension,$(GOOS)) \
		-a $(VERBOSE_GO) -tags "static_build netgo $(BUILDTAGS)" -installsuffix netgo \
		-ldflags "$(GO_LDFLAGS) -extldflags -static" $(GO_GCFLAGS) $(3);
endef

build-clean:
	rm -Rf $(PREFIX)/bin/*

build-x: ./cmd/machine.go
	$(foreach GOARCH,$(TARGET_ARCH),$(foreach GOOS,$(TARGET_OS),$(call gocross,$(GOOS),$(GOARCH),$<)))

$(PREFIX)/bin/docker-machine$(call extension,$(GOOS)): ./cmd/machine.go
	$(GO) build \
	-o $@ \
	$(VERBOSE_GO) -tags "$(BUILDTAGS)" \
	-ldflags "$(GO_LDFLAGS)" $(GO_GCFLAGS) $<

build: $(PREFIX)/bin/docker-machine

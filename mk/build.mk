build-clean:
	@rm -f $(PREFIX)/bin/*

# Cross builder helper
define gocross
	GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 go build -o $(PREFIX)/bin/$(1)-$(2)/docker-$(patsubst cmd/%.go,%,$3) \
		-a $(VERBOSE_GO) -tags "static_build netgo $(BUILDTAGS)" -installsuffix netgo -ldflags "$(GO_LDFLAGS) \
		-extldflags -static" $(GO_GCFLAGS) $(3);
endef

# XXX building with -a fails in debug (with -N -l) ????
$(PREFIX)/bin/docker-%: ./cmd/%.go $(shell find . -type f -name '*.go')
	$(GO) build -o $@ $(VERBOSE_GO) -tags "$(BUILDTAGS)" -ldflags "$(GO_LDFLAGS)" $(GO_GCFLAGS) $<

# Native build
build-simple: $(patsubst ./cmd/%.go,$(PREFIX)/bin/docker-%,$(wildcard ./cmd/*.go))

# Cross compilation targets
build-x-%: ./cmd/%.go $(shell find . -type f -name '*.go')
	@$(foreach GOARCH,$(TARGET_ARCH),$(foreach GOOS,$(TARGET_OS),$(call gocross,$(GOOS),$(GOARCH),$<)))

# Cross-build
build-x: $(patsubst ./cmd/%.go,build-x-%,$(wildcard ./cmd/*.go))
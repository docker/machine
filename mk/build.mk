build-clean:
	@rm -f $(PREFIX)/bin/*

# Simple build
build-simple: $(PREFIX)/bin/$(PKG_NAME)

# XXX building with -a fails in debug (with -N -l) ????
$(PREFIX)/bin/$(PKG_NAME): $(shell find . -type f -name '*.go')
	@go build -o $@ $(VERBOSE_GO) -tags "$(BUILDTAGS)" -ldflags "$(GO_LDFLAGS)" $(GO_GCFLAGS) ./main.go

# Cross-build: careful, does always rebuild!
build-x: clean
	$(if $(GOX), , \
		$(error Please install gox: go get -u github.com/mitchellh/gox))
	@$(GOX) \
		-os "$(TARGET_OS)" \
		-arch "$(TARGET_ARCH)" \
		-output="$(PREFIX)/bin/docker-machine_{{.OS}}-{{.Arch}}" \
		-ldflags="$(GO_LDFLAGS)" \
		-tags="$(BUILDTAGS)" \
		-gcflags="$(GO_GCFLAGS)" \
		-parallel=$(PARALLEL) \
		-rebuild $(VERBOSE_GOX)

# Cross builder helper
# define gocross
# 	GOOS=$(1) GOARCH=$(2) CGO_ENABLED=0 go build -o $(PREFIX)/$(PKG_NAME)_$(1)-$(2) \
# 		-a $(VERBOSE_GO) -tags "static_build netgo $(BUILDTAGS)" -installsuffix netgo -ldflags "$(GO_LDFLAGS) -extldflags -static" $(GO_GCFLAGS) ./main.go;
# endef

# Native build-x (no gox)
# build-x: $(shell find . -type f -name '*.go')
# 	@$(foreach GOARCH,$(TARGET_ARCH),$(foreach GOOS,$(TARGET_OS),$(call gocross,$(GS),$(GA))))
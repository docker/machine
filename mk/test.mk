# Unset docker-machine env vars, so they won't affect tests
unexport DOCKER_CERT_PATH
unexport DOCKER_HOST
unexport DOCKER_MACHINE_NAME
unexport DOCKER_TLS_VERIFY

# Quick test. You can bypass long tests using: `if testing.Short() { t.Skip("Skipping in short mode.") }`
test-short:
	$(GO) test $(VERBOSE_GO) -test.short -tags "$(BUILDTAGS)" $(PKGS)

# Runs long tests also, plus race detection
test-long:
	$(GO) test $(VERBOSE_GO) -race -tags "$(BUILDTAGS)" $(PKGS)

test-integration: build
	$(eval TESTSUITE=$(filter-out $@,$(MAKECMDGOALS)))
	test/integration/run-bats.sh $(TESTSUITE)

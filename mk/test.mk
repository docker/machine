# Quick test. You can bypass long tests using: `if testing.Short() { t.Skip("Skipping in short mode.") }`
test-short:
	@go test $(VERBOSE_GO) -test.short -tags "$(BUILDTAGS)" $(PKGS)

# Runs long tests also, plus race detection
test-long:
	@go test $(VERBOSE_GO) -race -tags "$(BUILDTAGS)" $(PKGS)

test-integration: build
	$(eval TESTSUITE=$(filter-out $@,$(MAKECMDGOALS)))
	test/integration/run-bats.sh $(TESTSUITE)

%:
	@:

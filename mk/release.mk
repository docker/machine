release-checksum:
	$(foreach MACHINE_FILE, $(wildcard $(PREFIX)/bin/*), \
		$(shell printf "%-50s %-50s\n" "sha256 $(shell basename $(MACHINE_FILE))" "$(shell openssl dgst -sha256 < $(MACHINE_FILE))" > /dev/stderr) \
		$(shell printf "%-50s %-50s\n" "md5 $(shell basename $(MACHINE_FILE))" "$(shell openssl dgst -md5 < $(MACHINE_FILE))" > /dev/stderr) \
		)
	@:

release: clean dco fmt test test-long build-x release-checksum
	# Github infos
	GH_USER ?= $(shell git config --get remote.origin.url | sed -e 's/.*[:/]\(.*\)\/\([^.]*\)\(.*\)/\1/')
	GH_REPO ?= $(shell git config --get remote.origin.url | sed -e 's/.*[:/]\(.*\)\/\([^.]*\)\(.*\)/\2/')

	$(if $(GITHUB_TOKEN), , \
		$(error GITHUB_TOKEN must be set for github-release))

	$(eval VERSION=$(filter-out $@,$(MAKECMDGOALS)))

	$(if $(VERSION), , \
		$(error Pass the version number as the first arg. E.g.: make release 1.2.3))

	git tag $(VERSION)
	git push --tags

	github-release release 
					--user $(GH_USER) \
					--repo $(GH_REPO) \
					--tag $(VERSION) \
					--name $(VERSION) \
					--description "" \
					--pre-release

	$(foreach MACHINE_FILE, $(wildcard $(PREFIX)/bin/*.zip), \
		$(shell github-release upload \
					--user $(GH_USER) \
					--repo $(GH_REPO) \
					--tag $(VERSION) \
					--name $(MACHINE_FILE) \
					--file $(MACHINE_FILE) \
			) \
		)

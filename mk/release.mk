release-checksum:
	$(foreach MACHINE_FILE, $(wildcard $(PREFIX)/bin/*.zip), \
		$(shell openssl dgst -sha256 < "$(MACHINE_FILE)" > "$(MACHINE_FILE).sha256" && \
						openssl dgst -md5 < "$(MACHINE_FILE)" > "$(MACHINE_FILE).md5" \
		))
	@:

release-pack:
	find ./bin -type d -mindepth 1 -exec zip -r -j {}.zip {} \;

release: clean dco fmt test test-long build-x release-pack release-checksum
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

%:
	@:

dep-save:
	$(if $(GODEP), , \
		$(error Please install godep: go get github.com/tools/godep))
	godep save $(shell go list ./... | grep -v vendor/)

dep-restore:
	$(if $(GODEP), , \
		$(error Please install godep: go get github.com/tools/godep))
	godep restore $(shell go list ./... | grep -v vendor/)

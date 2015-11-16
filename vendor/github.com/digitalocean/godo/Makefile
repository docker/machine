OPEN = $(shell which xdg-open || which gnome-open || which open)

cov:
	@@gocov test | gocov-html > /tmp/coverage.html
	@@${OPEN} /tmp/coverage.html

ci:
	go get -d -v -t ./...
	go build ./...
	go test -v ./...

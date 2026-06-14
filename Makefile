BINARY     := ccsm
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS    := -ldflags "-X main.version=$(VERSION)"
DIST       := dist

.PHONY: build install test test-verbose test-e2e test-all vet clean build-all release-dry

build:
	go build $(LDFLAGS) -o $(DIST)/$(BINARY) .

install:
	GOBIN=$${GOBIN:-/usr/local/bin} go install $(LDFLAGS) .

vet:
	go vet ./...

test:
	go test ./internal/...

test-verbose:
	go test -v ./internal/...

clean:
	rm -rf $(DIST)/

build-all:
	GOOS=linux  GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-amd64 .
	GOOS=linux  GOARCH=arm64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST)/$(BINARY)-darwin-arm64 .

test-e2e:
	go test -v -count=1 ./tests/...

test-all:
	go test -v -count=1 ./...

release-dry:
	goreleaser release --snapshot --clean

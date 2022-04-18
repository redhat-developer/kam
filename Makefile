BIN_DIR=bin
DIST_DIR=dist
EXECUTABLE=kam
WINDOWS=$(EXECUTABLE)_windows_amd64.exe
ARCHITECTURE=$(shell go env GOARCH)
LINUX=$(EXECUTABLE)_linux_$(ARCHITECTURE)
DARWIN=$(EXECUTABLE)_darwin_amd64
PKGS := $(shell go list  ./... | grep -v test/e2e | grep -v vendor)
FMTPKGS := $(shell go list  ./... | grep -v vendor)
VERSION=$(shell git describe --tags --always --long --dirty)
LD_FLAGS='-s -w -X github.com/redhat-developer/kam/pkg/cmd/version.Version=$(VERSION) -extldflags "-Wl,-z,now"'

.PHONY: all_platforms
all_platforms: windows linux darwin

.PHONY: windows
windows: $(WINDOWS)

.PHONY: linux
linux: $(LINUX)

.PHONY: darwin
darwin: $(DARWIN)

$(WINDOWS):
	env GOOS=windows GOARCH=amd64 go build -o $(DIST_DIR)/$(WINDOWS)  -ldflags=$(LD_FLAGS)  cmd/kam/kam.go

$(LINUX):
	env GOOS=linux GOARCH=$(ARCHITECTURE) go build -o $(DIST_DIR)/$(LINUX) -ldflags=$(LD_FLAGS) cmd/kam/kam.go

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/$(DARWIN) -ldflags=$(LD_FLAGS) cmd/kam/kam.go

default: bin

.PHONY: all
all:  gomod_tidy gofmt bin test

.PHONY: gomod_tidy
gomod_tidy:
	 go mod tidy

.PHONY: gofmt
gofmt:
	go fmt $(FMTPKGS)

.PHONY: bin
bin:
	go build -o $(BIN_DIR)/$(EXECUTABLE) -ldflags=$(LD_FLAGS) cmd/kam/kam.go

.PHONY: install
install:
	go install -v -ldflags=$(LD_FLAGS) cmd/kam/kam.go

.PHONY: test
test:
	 go test $(PKGS)

.PHONY: clean
clean:
	@rm -f $(DIST_DIR)/* $(BIN_DIR)/*

.PHONY: cmd-docs
cmd-docs:
	go run tools/cmd-docs/main.go

.PHONY: prepare-test-cluster
prepare-test-cluster:
	. ./scripts/prepare-test-cluster.sh

.PHONY: e2e
e2e:
GODOG_OPTS = --godog.tags=basic

e2e:
	@go test --timeout=180m ./test/e2e -v $(GODOG_OPTS)


.PHONY: e2e-local
e2e-local:
	@go test --timeout=180m ./test/e2e -v --godog.tags=local

.PHONY: checksum
checksum:
	. ./scripts/generate-checksum.sh
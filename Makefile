
EXECUTABLE=kam
WINDOWS=$(EXECUTABLE)_windows_amd64.exe
LINUX=$(EXECUTABLE)_linux_amd64
DARWIN=$(EXECUTABLE)_darwin_amd64
VERSION=$(shell git describe --tags --always --long --dirty)
LD_FLAGS="-s -w -X github.com/redhat-developer/kam/pkg/cmd/version.Version=$(VERSION)"

.PHONY: all_platforms
all_platforms: windows linux darwin 

.PHONY: windows
windows: $(WINDOWS)

.PHONY: linux
linux: $(LINUX)

.PHONY: darwin
darwin: $(DARWIN) 

$(WINDOWS):
	env GOOS=windows GOARCH=amd64 go build -i -v -o $(WINDOWS)  -ldflags=$(LD_FLAGS)  cmd/kam/kam.go

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -i -v -o $(LINUX)  -ldflags=$(LD_FLAGS) cmd/kam/kam.go

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -i -v -o $(DARWIN) -ldflags=$(LD_FLAGS) cmd/kam/kam.go	

default: bin

.PHONY: all
all:  gomod_tidy gofmt bin test

.PHONY: gomod_tidy
gomod_tidy:
	 go mod tidy

.PHONY: gofmt
gofmt:
	go fmt -x ./...

.PHONY: bin
bin:
	go build -i -v -o $(EXECUTABLE) -ldflags=$(LD_FLAGS) cmd/kam/kam.go 

.PHONY: install
install:
	go install -v -ldflags=$(LD_FLAGS) cmd/kam/kam.go

.PHONY: test
test:
	 go test ./...

.PHONY: clean
clean:
	@rm -f $(WINDOWS) $(LINUX) $(DARWIN) ${EXECUTABLE} 


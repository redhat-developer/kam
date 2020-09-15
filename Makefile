
EXECUTABLE=gitops
WINDOWS=$(EXECUTABLE)_windows_amd64.exe
LINUX=$(EXECUTABLE)_linux_amd64
DARWIN=$(EXECUTABLE)_darwin_amd64

.PHONY: all_platforms
all_platforms: windows linux darwin 

.PHONY: windows
windows: $(WINDOWS)

.PHONY: linux
linux: $(LINUX)

.PHONY: darwin
darwin: $(DARWIN) 

$(WINDOWS):
	env GOOS=windows GOARCH=amd64 go build -i -v -o $(WINDOWS) cmd/gitops-cli/gitops.go

$(LINUX):
	env GOOS=linux GOARCH=amd64 go build -i -v -o $(LINUX)  cmd/gitops-cli/gitops.go

$(DARWIN):
	env GOOS=darwin GOARCH=amd64 go build -i -v -o $(DARWIN) cmd/gitops-cli/gitops.go	

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
	go build cmd/gitops-cli/gitops.go 

.PHONY: install
install:
	go install ./cmd/gitops-cli/gitops.go

.PHONY: test
test:
	 go test ./...

.PHONY: clean
clean:
	@rm -f $(WINDOWS) $(LINUX) $(DARWIN) ${EXECUTABLE} 


FILES := gitops

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
	@rm -f ${FILES}


#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

export ARTIFACTS_DIR="/tmp/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACTS_DIR
export PATH=$PATH:$GOPATH/bin
# set location for golangci-lint cache
# otherwise /.cache is used, and it fails on permission denied
export GOLANGCI_LINT_CACHE="/tmp/.cache"

git describe --always --long --dirty
go version
go env
make gomod_tidy
make bin
make test

# crosscompile and publish artifacts
make all_platforms
pwd
ls -l
ls -l $GOPATH/src/github.com/redhat-developer/kam
#cp $GOPATH/src/github.com/redhat-developer/kam/kam_linux_amd64.exe $ARTIFACTS_DIR
#cp $GOPATH/src/github.com/redhat-developer/kam/kam_darwin_amd64.exe $ARTIFACTS_DIR
#cp $GOPATH/src/github.com/redhat-developer/kam/kam_windows_amd64.exe $ARTIFACTS_DIR


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
if [[ $(go fmt `go list ./... | grep -v vendor`) ]]; then
    echo "not well formatted sources are found"
    exit 1
fi
make gomod_tidy
if [[ ! -z $(git status -s) ]]
then
    echo "Go mod state is not clean."
    exit 1
fi
make test
make cmd-docs
if [[ ! -z $(git status -s) ]]
then
    echo "command-documentation is out of date (run make cmd-docs)."
    exit 1
else
    echo "command-documentation is up-to-date."
fi

# crosscompile and publish artifacts
make all_platforms

cp dist/kam_darwin_amd64 $CUSTOM_HOMEDIR
cp dist/kam_linux_amd64 $CUSTOM_HOMEDIR
cp dist/kam_windows_amd64.exe $CUSTOM_HOMEDIR

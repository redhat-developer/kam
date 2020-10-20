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
make test
make cmd-docs
if [[ ! -z $(git status -s) ]]
then
    echo "command-documentation is out of date (run make cmd-docs)."
    exit 1
else
    echo "command-documentation is up-to-date."
fi

make bin

# crosscompile and publish artifacts
make all_platforms

cp kam_darwin_amd64 $CUSTOM_HOMEDIR
cp kam_linux_amd64 $CUSTOM_HOMEDIR
cp kam_windows_amd64.exe $CUSTOM_HOMEDIR

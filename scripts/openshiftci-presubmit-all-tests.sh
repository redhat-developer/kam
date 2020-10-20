#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

export CI="prow"
make prepare-test-cluster
make bin
export PATH="$PATH:$(pwd)"
export ARTIFACTS_DIR="/tmp/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACTS_DIR

kam version

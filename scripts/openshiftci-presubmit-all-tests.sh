#!/bin/sh

# fail if some commands fails
set -e
# show commands
set -x

make bin
export PATH="$PATH:$(pwd)"
export ARTIFACTS_DIR="/tmp/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACTS_DIR

kam version


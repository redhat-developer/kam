#!/bin/sh

# fail if some commands fails
set -e

# Do not show token in CI log
set +x
export GITHUB_TOKEN=`cat $KAM_GITHUB_TOKEN_FILE`

# show commands
set -x
export CI="prow"
make prepare-test-cluster
make bin

INSTALL_ARGOCD="./scripts/install-argocd.sh"
sh $INSTALL_ARGOCD

INSTALL_DOCKER="./scripts/install-docker-cli.sh"
sh $INSTALL_DOCKER

export PATH="$PATH:$(pwd)/bin"
export ARTIFACTS_DIR="/tmp/artifacts"
export CUSTOM_HOMEDIR=$ARTIFACTS_DIR

# Copy kubeconfig to temporary kubeconfig file and grant
# read and Write permission to temporary kubeconfig file
TMP_DIR=$(mktemp -d)
cp $KUBECONFIG $TMP_DIR/kubeconfig
chmod 640 $TMP_DIR/kubeconfig
export KUBECONFIG=$TMP_DIR/kubeconfig

# Login as developer
oc login -u developer -p developer

# Check login user name for debugging purpose
oc whoami
login_user=`oc whoami`
if [[ $login_user == *"developer"* ]]; then
    echo "Login to the cluster as a developer user"
else
    echo "Fail to login as a developer user"
    exit 1
fi

# assert that kam is on the path
kam version

# Run e2e test
make e2e

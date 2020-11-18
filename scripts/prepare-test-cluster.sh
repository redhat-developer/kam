#!/bin/bash
set -x
SETUP_OPERATORS="./scripts/setup-operators.sh"
# Overrideable information
DEFAULT_INSTALLER_ASSETS_DIR=${DEFAULT_INSTALLER_ASSETS_DIR:-$(pwd)}
KUBEADMIN_USER=${KUBEADMIN_USER:-"kubeadmin"}
KUBEADMIN_PASSWORD_FILE=${KUBEADMIN_PASSWORD_FILE:-"${DEFAULT_INSTALLER_ASSETS_DIR}/auth/kubeadmin-password"}
# Exported to current env
ORIGINAL_KUBECONFIG=${KUBECONFIG:-"${DEFAULT_INSTALLER_ASSETS_DIR}/auth/kubeconfig"}
export KUBECONFIG=$ORIGINAL_KUBECONFIG

# list of namespace to create
OPERATOR_NAMESPACES="cicd argocd"

# Attempt resolution of kubeadmin, only if a CI is not set
if [ -z $CI ]; then
    # Check if nessasary files exist
    if [ ! -f $KUBEADMIN_PASSWORD_FILE ]; then
        echo "Could not find kubeadmin password file"
        exit 1
    fi

    if [ ! -f $KUBECONFIG ]; then
        echo "Could not find kubeconfig file"
        exit 1
    fi

    # Get kubeadmin password from file
    KUBEADMIN_PASSWORD=`cat $KUBEADMIN_PASSWORD_FILE`

    # Login as admin user
    oc login -u $KUBEADMIN_USER -p $KUBEADMIN_PASSWORD
else
    # Copy kubeconfig to temporary kubeconfig file
    # Read and Write permission to temporary kubeconfig file
    TMP_DIR=$(mktemp -d)
    cp $KUBECONFIG $TMP_DIR/kubeconfig
    chmod 640 $TMP_DIR/kubeconfig
    export KUBECONFIG=$TMP_DIR/kubeconfig
fi

# Create the namespace for operator installation namespace
for i in `echo $OPERATOR_NAMESPACES`; do
    # create the namespace
    oc new-project $i
done

# Setup the cluster for sealed secrets, pipelines and argocd operator
sh $SETUP_OPERATORS

# Client version
oc version

# Project list
oc projects

# KUBECONFIG cleanup only if CI is set
if [ ! -f $CI ]; then
    rm -rf $KUBECONFIG
    export KUBECONFIG=$ORIGINAL_KUBECONFIG
fi

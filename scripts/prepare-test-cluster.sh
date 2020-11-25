#!/bin/bash
set -x
# Setup to find necessary data from cluster setup
# Constants
HTPASSWD_FILE="./htpass"
USERPASS="developer"
HTPASSWD_SECRET="htpasswd-secret"
SETUP_OPERATORS="./scripts/setup-operators.sh"
# Overrideable information
DEFAULT_INSTALLER_ASSETS_DIR=${DEFAULT_INSTALLER_ASSETS_DIR:-$(pwd)}
KUBEADMIN_USER=${KUBEADMIN_USER:-"kubeadmin"}
KUBEADMIN_PASSWORD_FILE=${KUBEADMIN_PASSWORD_FILE:-"${DEFAULT_INSTALLER_ASSETS_DIR}/auth/kubeadmin-password"}
# Default values
OC_LOGIN_SUCCEEDED="false"
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
# else
#     # Copy kubeconfig to temporary kubeconfig file
#     # Read and Write permission to temporary kubeconfig file
#     TMP_DIR=$(mktemp -d)
#     cp $KUBECONFIG $TMP_DIR/kubeconfig
#     chmod 640 $TMP_DIR/kubeconfig
#     export KUBECONFIG=$TMP_DIR/kubeconfig
fi

# Create the namespace for operator installation namespace
for i in `echo $OPERATOR_NAMESPACES`; do
    # create the namespace
    oc new-project $i
    # # Let developer user have access to the project
    # oc adm policy add-role-to-user edit developer
done

# Setup the cluster for sealed secrets, pipelines and argocd operator
sh $SETUP_OPERATORS

# # Remove existing htpasswd file, if any
# if [ -f $HTPASSWD_FILE ]; then
#     rm -rf $HTPASSWD_FILE
# fi

# # Set so first time -c parameter gets applied to htpasswd
# HTPASSWD_CREATED=" -c "

# # Create htpasswd entries for developer
# htpasswd -b $HTPASSWD_CREATED $HTPASSWD_FILE developer $USERPASS
# HTPASSWD_CREATED=""

# # Create secret in cluster and replace
# oc create secret generic ${HTPASSWD_SECRET} --from-file=htpasswd=${HTPASSWD_FILE} -n openshift-config --dry-run=client -o yaml | oc apply -f -

# # Upload htpasswd as new login config
# oc apply -f - <<EOF
# apiVersion: config.openshift.io/v1
# kind: OAuth
# metadata:
#   name: cluster
# spec:
#   identityProviders:
#   - name: htpassidp1
#     challenge: true
#     login: true
#     mappingMethod: claim
#     type: HTPasswd
#     htpasswd:
#       fileData:
#         name: ${HTPASSWD_SECRET}
# EOF

# # Login as developer and check for stable server
# for i in {1..40}; do
#     # Try logging in as developer
#     oc login -u developer -p $USERPASS &> /dev/null
#     if [ $? -eq 0 ]; then
#         # If login succeeds, assume success
# 	    OC_LOGIN_SUCCEEDED="true"
#         # Attempt failure of `oc whoami`
#         for j in {1..25}; do
#             oc whoami &> /dev/null
#             if [ $? -ne 0 ]; then
#                 # If `oc whoami` fails, assume fail and break out of trying `oc whoami`
#                 OC_LOGIN_SUCCEEDED="false"
#                 break
#             fi
#             sleep 2
#         done
#         # If `oc whoami` never failed, break out trying to login again
#         if [ $OC_LOGIN_SUCCEEDED == "true" ]; then
#             break
#         fi
#     fi
#     sleep 3
# done

# if [ $OC_LOGIN_SUCCEEDED == "false" ]; then
#     echo "Failed to login as developer"
#     exit 1
# fi

# Client version
oc version

# Project list
oc projects

# # KUBECONFIG cleanup only if CI is set
# if [ ! -f $CI ]; then
#     rm -rf $KUBECONFIG
#     export KUBECONFIG=$ORIGINAL_KUBECONFIG
# fi

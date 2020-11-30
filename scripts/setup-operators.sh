#!/bin/bash
set -x
echo "Starting sealed-secrets operator installation"
install_sealed_secrets_operator(){
# If the Operator uses the SingleNamespace mode for example cicd namespace
# and you do not already have an appropriate OperatorGroup in place,
# you must create one.
# Check for details - https://docs.openshift.com/container-platform/4.5/operators/admin/olm-adding-operators-to-cluster.html#olm-installing-operator-from-operatorhub-using-cli_olm-adding-operators-to-a-cluster
oc create -f - <<EOF
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  generateName: cicd-
  namespace: cicd
spec:
  targetNamespaces:
  - cicd
EOF

# Apply the sealed-secrets operator subscription
oc create -f - <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: sealed-secrets-operator-helm
  namespace: cicd
spec:
  channel: alpha
  name: sealed-secrets-operator-helm
  source: community-operators
  sourceNamespace: openshift-marketplace
EOF
}

install_sealed_secrets_operator
# sealed-secrets operator status check
count=0
while [ "$count" -lt "5" ];
do
    operator_status=`oc get csv -n cicd | grep sealed-secrets-operator`
    if [[ $operator_status == *"Succeeded"* ]]; then
        break
    else
        count=`expr $count + 1`
        sleep 10
    fi
done
echo "Completed sealed-secrets operator installation"

echo "Starting sealed-secrets operator instance creation"
create_sealed_secrets_operator_instance(){
oc create -f - <<EOF
apiVersion: bitnami.com/v1alpha1
kind: SealedSecretController
metadata: 
  name: sealedsecretscontroller
  namespace: cicd
spec: 
  affinity: {}
  nodeSelector: {}
  securityContext: 
    fsGroup: ""
    runAsUser: ""
  serviceAccount: 
    name: ""
  tolerations: []
EOF
}
create_sealed_secrets_operator_instance
echo "Completed sealed-secrets operator instance creation"

echo "Starting openshift-pipelines operator installation"
install_openshift_pipelines_operator() {
oc create -f - <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: openshift-pipelines-operator-rh
  namespace: openshift-operators
spec:
  channel: ocp-4.5
  name: openshift-pipelines-operator-rh
  source: redhat-operators
  sourceNamespace: openshift-marketplace
EOF
}

install_openshift_pipelines_operator
# pipelines operator status
count=0
while [ "$count" -lt "5" ];
do
    operator_status=`oc get csv -n openshift-operators | grep openshift-pipelines-operator`
    if [[ $operator_status == *"Succeeded"* ]]; then
        break
    else
        count=`expr $count + 1`
        sleep 10
    fi
done
echo "Completed openshift-pipelines operator installation"

echo "Starting argocd operator installation"
install_argocd_operator(){
# If the Operator uses the SingleNamespace mode for example argocd namespace
# and you do not already have an appropriate OperatorGroup in place,
# you must create one.
# Check for details - https://docs.openshift.com/container-platform/4.5/operators/admin/olm-adding-operators-to-cluster.html#olm-installing-operator-from-operatorhub-using-cli_olm-adding-operators-to-a-cluster
oc create -f - <<EOF
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  generateName: argocd-
  namespace: argocd
spec:
  targetNamespaces:
  - argocd
EOF

# Apply the argocd operator subscription
oc create -f - <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata: 
  labels: 
    operators.coreos.com/argocd-operator.argocd: ""
  name: argocd-operator
  namespace: argocd
spec: 
  channel: alpha
  name: argocd-operator
  source: community-operators
  sourceNamespace: openshift-marketplace
EOF
}

install_argocd_operator
# argocd operator status check
count=0
while [ "$count" -lt "5" ];
do
    operator_status=`oc get csv -n argocd | grep argocd-operator`
    if [[ $operator_status == *"Succeeded"* ]]; then
        break
    else
        count=`expr $count + 1`
        sleep 10
    fi
done
echo "Completed argocd operator installation"

# Due to an open issue https://github.com/argoproj-labs/argocd-operator/issues/107
# the operator may not create enough privileges to manage multiple namespaces.
# In order to solve this apply:
echo "Add Role Binding"
oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:argocd:argocd-application-controller

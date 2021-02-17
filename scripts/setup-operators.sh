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
while [ "$count" -lt "12" ];
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
  name: sealedsecretcontroller
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

echo "Starting OpenShift GitOps operator installation"
install_openshift_gitops_operator(){
oc create -f - <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: openshift-gitops-operator
  namespace: openshift-operators
  labels:
    operators.coreos.com/openshift-gitops-operator.openshift-operators: ''
spec:
  channel: preview
  installPlanApproval: Automatic
  name: openshift-gitops-operator
  source: redhat-operators
  sourceNamespace: openshift-marketplace
EOF
}

install_openshift_gitops_operator
# GitOps operator status check
count=0
while [ "$count" -lt "5" ];
do
    operator_status=`oc get csv -n openshift-operators | grep openshift-gitops-operator`
    if [[ $operator_status == *"Succeeded"* ]]; then
        break
    else
        count=`expr $count + 1`
        sleep 10
    fi
done
echo "Completed OpenShift GitOps operator installation"

echo "Provide cluster-admin access to argocd-application-controller service account"
oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:openshift-gitops:argocd-cluster-argocd-application-controller

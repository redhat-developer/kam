#!/bin/bash
set -x
echo "Starting openshift-pipelines operator installation"
install_openshift_pipelines_operator() {
oc create -f - <<EOF
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: openshift-pipelines-operator-rh
  namespace: openshift-operators
  labels:
    operators.coreos.com/openshift-pipelines-operator-rh.openshift-operators: ''
spec:
  channel: stable
  name: openshift-pipelines-operator-rh
  source: redhat-operators
  sourceNamespace: openshift-marketplace
EOF
}

install_openshift_pipelines_operator
# pipelines operator status
count=0
while [ "$count" -lt "15" ];
do
    operator_status=`oc get csv -n openshift-operators | grep redhat-openshift-pipelines`
    if [[ $operator_status == *"Succeeded"* ]]; then
        break
    else
        count=`expr $count + 1`
        sleep 10
    fi
done
echo "Completed openshift-pipelines operator installation"

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
  channel: stable
  installPlanApproval: Automatic
  name: openshift-gitops-operator
  source: redhat-operators
  sourceNamespace: openshift-marketplace
EOF
}

install_openshift_gitops_operator
# GitOps operator status check
count=0
while [ "$count" -lt "15" ];
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
oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:openshift-gitops:openshift-gitops-argocd-application-controller

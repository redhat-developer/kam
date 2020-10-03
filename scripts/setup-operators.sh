#!/bin/bash
set -x
echo "starting"
install_sealed_secrets_operator(){
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata: 
    labels: 
      operators.coreos.com/sealed-secrets-operator-helm.cicd: ""
    name: sealed-secrets-operator-helm
    namespace: cicd
  spec: 
    channel: alpha
    installPlanApproval: Automatic
    name: sealed-secrets-operator-helm
    source: community-operators
    sourceNamespace: openshift-marketplace
    startingCSV: sealed-secrets-operator-helm.v0.0.2
EOF
}

install_sealed_secrets_operator
# sealed secrets operator status check
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

install_openshift_pipelines_operator() {
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata: 
    labels: 
      operators.coreos.com/openshift-pipelines-operator-rh.openshift-operators: ""
    name: openshift-pipelines-operator-rh
    namespace: openshift-operators
  spec: 
    channel: ocp-4.5
    installPlanApproval: Automatic
    name: openshift-pipelines-operator-rh
    source: redhat-operators
    sourceNamespace: openshift-marketplace
    startingCSV: openshift-pipelines-operator.v1.0.1
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

install_argocd_operator(){
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
    installPlanApproval: Automatic
    name: argocd-operator
    source: community-operators
    sourceNamespace: openshift-marketplace
    startingCSV: argocd-operator.v0.0.13
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

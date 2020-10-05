#!/bin/bash
set -x
echo "Starting sealed secrets operator installation"
install_sealed_secrets_operator(){
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata: 
    labels: 
      operators.coreos.com/sealed-secrets-operator-helm.cicd: ""
    managedFields: 
      - 
        apiVersion: operators.coreos.com/v1alpha1
        manager: olm
        operation: Update
      - 
        apiVersion: operators.coreos.com/v1alpha1
        manager: catalog
        operation: Update
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
echo "Completes sealed secrets operator installation"

# create_sealed_secrets_operator_instance(){
#   oc create -f - <<EOF
#   apiVersion: bitnami.com/v1alpha1
#   kind: SealedSecretController
#   metadata: 
#     finalizers: 
#       - uninstall-helm-release
#     managedFields: 
#       - 
#         apiVersion: bitnami.com/v1alpha1
#         manager: oc
#         operation: Update
#       - 
#         apiVersion: bitnami.com/v1alpha1
#         manager: helm-operator
#         operation: Update
#     name: sealedsecretsontroller
#     namespace: cicd
#   spec: 
#     affinity: {}
#     controller: 
#       create: true
#     crd: 
#       create: true
#       keep: true
#     image: 
#       pullPolicy: IfNotPresent
#       repository: "quay.io/bitnami/sealed-secrets-controller@sha256:8e9a37bb2e1a6f3a8bee949e3af0e9dab0d7dca618f1a63048dc541b5d554985"
#     ingress: 
#       annotations: {}
#       enabled: false
#       hosts: 
#         - chart-example.local
#       path: /v1/cert.pem
#       tls: []
#     networkPolicy: false
#     nodeSelector: {}
#     podAnnotations: {}
#     podLabels: {}
#     priorityClassName: ""
#     rbac: 
#       create: true
#       pspEnabled: false
#     resources: {}
#     secretName: sealed-secrets-key
#     securityContext: 
#       fsGroup: ""
#       runAsUser: ""
#     serviceAccount: 
#       create: true
#       name: ""
#     tolerations: []
# EOF
# }

echo "Starting openshift pipelines operator installation"
install_openshift_pipelines_operator() {
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata: 
    labels: 
      operators.coreos.com/openshift-pipelines-operator-rh.openshift-operators: ""
    managedFields: 
      - 
        apiVersion: operators.coreos.com/v1alpha1
        manager: olm
        operation: Update
      - 
        apiVersion: operators.coreos.com/v1alpha1
        manager: catalog
        operation: Update
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
echo "Completes openshift pipelines operator installation"

echo "Starting argocd operator installation"
install_argocd_operator(){
  oc create -f - <<EOF
  apiVersion: operators.coreos.com/v1alpha1
  kind: Subscription
  metadata: 
    labels: 
      operators.coreos.com/argocd-operator.argocd: ""
    managedFields: 
      - 
        apiVersion: operators.coreos.com/v1alpha1
        manager: olm
        operation: Update
      - 
        apiVersion: operators.coreos.com/v1alpha1
        manager: catalog
        operation: Update
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
echo "Completes argocd operator installation"

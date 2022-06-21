# #!/bin/bash
set -x

echo "Installing OpenShift Pipelines operator"
echo -e "Ensure pipelines subscription exists"
oc get subscription openshift-pipelines-operator-rh -n openshift-operators 2>/dev/null || \
cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: openshift-pipelines-operator-rh
  namespace: openshift-operators
spec:
  channel: stable
  name: openshift-pipelines-operator-rh
  source: redhat-operators
  sourceNamespace: openshift-marketplace
EOF

for i in {1..150}; do  # timeout after 5 minutes
  pods="$(oc get pods -n openshift-operators --no-headers 2>/dev/null | wc -l)"
  if [[ "${pods}" -ge 1 ]]; then
    echo -e "\nWaiting for Pipelines operator pod"
    oc wait --for=condition=Ready -n openshift-operators -l name=openshift-pipelines-operator pod --timeout=5m
    retval=$?
    if [[ "${retval}" -gt 0 ]]; then exit "${retval}"; else break; fi
  fi
  if [[ "${i}" -eq 150 ]]; then
    echo "Timeout: pod was not created."
    exit 2
  fi  
  echo -n "."
  sleep 2
done

for i in {1..150}; do  # timeout after 5 minutes
  pods="$(oc get pods -n openshift-pipelines --no-headers 2>/dev/null | wc -l)"
  if [[ "${pods}" -ge 4 ]]; then
    echo -e "\nWaiting for Pipelines and Triggers pods"
    oc wait --for=condition=Ready -n openshift-pipelines pod --timeout=5m \
      -l 'app in (tekton-pipelines-controller,tekton-pipelines-webhook,tekton-triggers-controller,
      tekton-triggers-webhook)'
    retval=$?
    if [[ "${retval}" -gt 0 ]]; then exit "${retval}"; else break; fi
  fi
  if [[ "${i}" -eq 150 ]]; then
    echo "Timeout: pod was not created."
    exit 2
  fi  
  echo -n "."
  sleep 2
done

echo "Installing OpenShift GitOps operator"
echo -e "Ensure gitops subscription exists"
oc get subscription openshift-gitops-operator-rh -n openshift-operators 2>/dev/null || \
cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: openshift-gitops-operator
  namespace: openshift-operators
spec:
  channel: stable
  installPlanApproval: Automatic
  name: openshift-gitops-operator
  source: redhat-operators
  sourceNamespace: openshift-marketplace
EOF

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
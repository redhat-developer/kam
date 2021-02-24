## Install OpenShift GitOps operator

Install OpenShift GitOps operator from OperatorHub which installs the following components

1. Argo CD instance in `openshift-gitops` namespace
2. OpenShift Pipelines Operator
3. GitOps Service in `openshift-gitops` namespace (`openshift-pipelines-app-delivery` in case of a 4.6 cluster)

Follow the installation wizard and deploy the operator with defaults.

![screenshot](img/gitops-listing.png)

![screenshot](img/gitops-installation.png)

![screenshot](img/argocd-instance.png)

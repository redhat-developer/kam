# ArgoCD Installation - with operator (Recommended)

There are three steps in this installation
* [create `argocd` namespace ](#Create-argocd-namespace)
* [install ArgoCD Operator](#install-ArgoCD-Operator)
* [Add Role Binding](#Add-Role-Binding)

## Create argocd namespace
Create argocd namespace to install the operator:

```shell
$ oc create namespace argocd
```

## Install ArgoCD Operator
Click on the ArgoCD operator as shown below in the OperatorHub on your OpenShift console and install the operator in the argocd namespace.


![screenshot](img/argocd-1.png)

![screenshot](img/argocd-2.png)

![screenshot](img/argocd-3.png)

![screenshot](img/argocd-4.png)

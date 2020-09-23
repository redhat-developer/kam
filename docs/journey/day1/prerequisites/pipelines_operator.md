## Install OpenShift Pipelines Operator

Login to OpenShift DevConsole and install Tekton Operator

![OpenShift Pipelines Operator](img/tekton-operator-install.gif)


## Install the Buildah ClusterTask

```shell
$ oc replace -f https://github.com/redhat-developer/gitops-cli/blob/master/docs/helpers/buildah.yaml
```

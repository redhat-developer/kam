## Install OpenShift Pipelines Operator

Login to OpenShift DevConsole and install Tekton Operator

![OpenShift Pipelines Operator](img/pipelines-operator-install.gif)


## Install the Buildah ClusterTask

```shell
$ oc replace -f https://raw.githubusercontent.com/redhat-developer/kam/master/docs/updates/buildah.yaml
```

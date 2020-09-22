## Install OpenShift Pipelines Operator

Login to OpenShift DevConsole and install Tekton Operator

![](img/tekton-operator-install.gif)


## Install the Buildah ClusterTask

```
oc replace -f https://gist.githubusercontent.com/sbose78/3e4fe119489bb0839c376dc9a9c603a5/raw/15424ea4df6c339f8238c132cf5ac33a94e8efc8/buildah.yaml
```

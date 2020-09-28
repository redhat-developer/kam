# GitOps Setup on DevConsole


1. Setup a GitOps pipeline on your cluster
    * [setup up sample GitOps pipelines](./setup-gitops.md)
2. Install GitOps service operator from Operator hub in all-namespaces.
3. Go to DevConsole
4. Application Stages nav-item should now be visible (feature-flagged on the availability of the operator)
5. Create a namespace - 
```shell
$ kubectl create namespace pipelines-{console_username}-github
```
6. Create secret with your github token in above namespace
```shell
$ kubectl create secret -n pipelines-{console_username}-github generic {console_username}-github-token --from-literal=token={user-token}
```
7. Application Stages should now be populated with the list of applications 
8. Clicking on an application will take you to the application details page


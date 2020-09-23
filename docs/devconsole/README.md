# GitOps Setup on DevConsole


1. Setup a GitOps pipeline on your cluster
    * Using ODO - https://gist.github.com/rohitkrai03/c01a93739dcfad8300d1e4e6450caba6
    * Using GitOps-CLI - [Docs not available] Contact gitops teams
2. Install GitOps service operator from Operator hub in all-namespaces (cannot be installed in a specific ns because of an ongoing issue https://coreos.slack.com/archives/CMP95ST2N/p1600695524017600)
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


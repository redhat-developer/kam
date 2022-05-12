# GitOps Setup on DevConsole

Visualize your GitOps environment stages and applications on DevConsole in OpenShift 4.9


1. Follow the [Day 1 Operations](../journey/day1) and make sure your Argo CD applications are in sync.

![Applications](../journey/day1/img/ArgoCD_UI.png)

2. Install the GitOps service operator from OperatorHub in all-namespaces.

![Operator](./img/GitOps_Operator.jpg)

![Installation](./img/GitOps_Installation.jpg)

3. The operator will install the GitOps backend service in `openshift-gitops` namespace. `Environments` nav-item will be visible under the Developer perspective on the availability of the operator.

![Service](./img/GitOps_Service.png)

4. Select the `Environments` nav-item. The Environments page will immediately show the list of applications.

![Application_List](./img/GitOps_4.9_Environments.png)

5. Hover over the icons under the Environment status column to see the sync status of all the environments.

![Application_List](./img/GitOps_4.8_Environments_Tooltip.png)

6. Clicking on the application name will take you to the application details page with the `Overview` tab selected.

![GitOps Details Page 4.11](./img/GitOps_4.11_App_Details_Overview.png)

7. Click on the `Deployment history` tab to get the successful deployment history for the application environments. This feature is new to Red Hat OpenShift 4.11.

![GitOps Deployment History Tab 4.11](./img/GitOps_4.11_App_Details_DeploymentHistory.png)

8. The history will update once a successful deployment or sync is made in Argo CD.

![GitOps Deployment History Tab 4.11](./img/GitOps_4.11_App_Details_DeploymentHistory_AfterSync.png)

9. Hover over the icons on the right of a resource to see it's health status and sync status. Only when a resource is out of sync or/and not healthy the corresponding icon is displayed.

![GitOps Details Page - Resources](./img/GitOps_4.9_Details_Resources.png)

Follow [Day 2 Operations](../journey/day2) to add more environments and applications to your GitOps workflow.
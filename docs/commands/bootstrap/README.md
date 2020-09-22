# Gitops Bootstrap Command

The `bootstrap` sub-command creates default environments for your initial application.

It outputs resource files in YAML format, Kustomization files, and a pipelines configuration file.

The following resources are written to filesystem.
   
* CI/CD environments with pipelines and resources
* ArgoCD environment
* Dev environment with an application/service
* Stage environment

```shell
$ gitops bootstrap
  --gitops-repo-url
  --service-repo-url
  --image-repo
  --dockercfgjson
  [--image-repo-internal-registry-hostname]
  [--gitops-webhook-secret]
  [--service-webhook-secret]
  [--prefix]
  [--output]
  [--overwrite]
  [--sealed-secrets-ns]
  [--sealed-secrets-svc]
  [--git-host-access-token]
  [--private-repo-driver]
  [--commit-status-tracker]
```

The directory layout generated is shown below.

```
.
├── config
│   ├── argocd
│   │   ├── argo-app.yaml
│   │   ├── argocd.yaml
│   │   ├── cicd-app.yaml
│   │   ├── dev-app-taxi-app.yaml
│   │   └── kustomization.yaml
│   └── cicd
│       ├── base
│       │   ├── 01-namespaces
│       │   │   └── cicd-environment.yaml
│       │   ├── 02-rolebindings
│       │   │   ├── pipeline-service-account.yaml
│       │   │   ├── pipeline-service-role.yaml
│       │   │   └── pipeline-service-rolebinding.yaml
│       │   ├── 03-secrets
│       │   │   ├── docker-config.yaml
│       │   │   ├── gitops-webhook-secret.yaml
│       │   │   └── webhook-secret-dev-taxi.yaml
│       │   ├── 04-tasks
│       │   │   ├── deploy-from-source-task.yaml
│       │   │   └── deploy-using-kubectl-task.yaml
│       │   ├── 05-pipelines
│       │   │   ├── app-ci-pipeline.yaml
│       │   │   └── ci-dryrun-from-push-pipeline.yaml
│       │   ├── 06-bindings
│       │   │   ├── dev-app-taxi-taxi-binding.yaml
│       │   │   └── gitlab-push-binding.yaml
│       │   ├── 07-templates
│       │   │   ├── app-ci-build-from-push-template.yaml
│       │   │   └── ci-dryrun-from-push-template.yaml
│       │   ├── 08-eventlisteners
│       │   │   └── cicd-event-listener.yaml
│       │   ├── 09-routes
│       │   │   └── gitops-webhook-event-listener.yaml
│       │   └── kustomization.yaml
│       └── overlays
│           └── kustomization.yaml
├── environments
│   ├── dev
│   │   ├── apps
│   │   │   └── app-taxi
│   │   │       ├── base
│   │   │       │   └── kustomization.yaml
│   │   │       ├── kustomization.yaml
│   │   │       ├── overlays
│   │   │       │   └── kustomization.yaml
│   │   │       └── services
│   │   │           └── taxi
│   │   │               ├── base
│   │   │               │   ├── config
│   │   │               │   │   ├── 100-deployment.yaml
│   │   │               │   │   ├── 200-service.yaml
│   │   │               │   │   └── kustomization.yaml
│   │   │               │   └── kustomization.yaml
│   │   │               ├── kustomization.yaml
│   │   │               └── overlays
│   │   │                   └── kustomization.yaml
│   │   └── env
│   │       ├── base
│   │       │   ├── dev-environment.yaml
│   │       │   ├── dev-rolebinding.yaml
│   │       │   └── kustomization.yaml
│   │       └── overlays
│   │           └── kustomization.yaml
│   └── stage
│       └── env
│           ├── base
│           │   ├── kustomization.yaml
│           │   └── stage-environment.yaml
│           └── overlays
│               └── kustomization.yaml
└── pipelines.yaml
```

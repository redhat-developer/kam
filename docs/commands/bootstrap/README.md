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
  [--internal-registry-hostname]
  [--gitops-webhook-secret]
  [--service-webhook-secret]
  [--sealed-secrets-ns]
  [--prefix]
  [--output]
  [--overwrite]
  [--sealed-secrets-ns]
  [--sealed-secrets-svc]
  [--status-tracker-access-token]
  [--private-repo-driver]
```

| Flag                                  | Description |
| ------------------------------------- | ----------- |
| --dockercfgjson                       | Filepath to config.json which authenticates the image push to the desired image registry. |
| --gitops-repo-url                     | Provide the URL for your GitOps repository e.g. https://github.com/organisation/repository.git |
| --gitops-webhook-secret               | Optional. Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the GitOps repository. (if not provided, it will be auto-generated)|
| --help                                | Help for bootstrap flags. |
| --image-repo                          | Image repository of the form <registry>/<username>/<repository> or <project>/<app> which is used to push newly built images. |
| --internal-registry-hostname          | Host-name for internal image registry e.g. docker-registry.default.svc.cluster.local:5000, used if you are pushing your images to the internal image registry |
| --output                              | Path to write GitOps resources (default ".") |
| --prefix                              | Add a prefix to the environment names(Dev, stage,prod,cicd etc.) to distinguish and identify individual environments. |
| --service-repo-url                    | Provide the URL for your Sevice repository e.g. https://github.com/organisation/repository.git which is source code to your first application. |
| --service-webhook-secret              | Optional. Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the GitOps repository. (if not provided, it will be auto-generated)|
| --overwrite                           | Optional. Overwrites previously existing GitOps configuration (if any) (default false) |
| --sealed-secrets-ns string            | Optional. Namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator (default "cicd") |
| --sealed-secrets-svc string           | Optional. Name of the Sealed Secrets Services that encrypts secrets (default "sealedsecretcontroller-sealed-secrets"") |
| --status-tracker-access-token string  | Optional. Used to authenticate requests to push commit-statuses to your Git hosting service|
| --private-repo-driver string          | Optional. If your Enterprise Git repositories are on a custom domain, please indicate which driver to use github or gitlab|

The following [directory layout](output) is generated.

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

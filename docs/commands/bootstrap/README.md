# KAM Bootstrap Command

The `bootstrap` sub-command creates default environments for your initial application.

It outputs resource files in YAML format, Kustomization files, and a pipelines configuration file.

The following resources are written to filesystem.

* CI/CD environments with pipelines and resources
* ArgoCD environment
* Dev environment with an application/service
* Stage environment

```
Bootstrap GitOps CI/CD Manifests

Usage:
  kam bootstrap [flags]

Examples:
  # Bootstrap OpenShift pipelines.
  kam bootstrap

Flags:
      --commit-status-tracker                          Enable or disable the commit-status-tracker which reports the success/failure of your pipelineruns to GitHub/GitLab (default true)
      --dockercfgjson string                           Filepath to config.json which authenticates the image push to the desired image registry  (default "~/.docker/config.json")
      --git-host-access-token string                   Used to authenticate repository clones, and commit-status notifications (if enabled)
      --gitops-repo-url string                         Provide the URL for your GitOps repository e.g. https://github.com/organisation/repository.git
      --gitops-webhook-secret string                   Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the GitOps repository. (if not provided, it will be auto-generated)
  -h, --help                                           help for bootstrap
      --image-repo string                              Image repository of the form <registry>/<username>/<repository> or <project>/<app> which is used to push newly built images
      --image-repo-internal-registry-hostname string   Host-name for internal image registry e.g. docker-registry.default.svc.cluster.local:5000, used if you are pushing your images to the internal image registry (default "image-registry.openshift-image-registry.svc:5000")
      --output string                                  Path to write GitOps resources (default ".")
      --overwrite                                      Overwrites previously existing GitOps configuration (if any)
  -p, --prefix string                                  Add a prefix to the environment names(Dev, stage,prod,cicd etc.) to distinguish and identify individual environments
      --private-repo-driver string                     If your Git repositories are on a custom domain, please indicate which driver to use github or gitlab
      --sealed-secrets-ns string                       Namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator (default "cicd")
      --sealed-secrets-svc string                      Name of the Sealed Secrets Services that encrypts secrets (default "sealedsecretcontroller-sealed-secrets")
      --service-repo-url string                        Provide the URL for your Service repository e.g. https://github.com/organisation/service.git
      --service-webhook-secret string                  Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the Service repository. (if not provided, it will be auto-generated)
  ```


## Flag mode

```shell
$ kam bootstrap
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

## Interactive mode

Running the bootstrap command without flags will trigger an interactive prompt. Each question in the prompt is accompanied by help, providing a brief explanation for the question. 

[![asciicast](https://asciinema.org/a/P3hsAu34gvYrxp6DPsA3AgWyn.svg)](https://asciinema.org/a/P3hsAu34gvYrxp6DPsA3AgWyn)

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

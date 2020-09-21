# Gitops Service Command

## Service add

The `service add` sub-command adds a new service to an existing environment.

Services are logically grouped by applications.

This command will create an application for the new service if the target application does not exist.  It outputs resources YAML files, Kustomization files, and updated Manifest to filesystem.

**NOTE**: Service deployment resources are not generated.  They must be manually added to `environments/<env-new>/ services/<service-name>/base/config` and update the Kustomization file `environments/<env-new>/ services/<service-name>/base/kustomization.yaml`

```shell
$ gitops service add
    --env-name
    --app-name
    --service-name
    [--git-repo-url]
    [--sealed-secrets-ns]
    [--webhook-secret]
    [--image-repo]
    [--internal-registry-hostname]
    [--pipelines-file]
```

| Flag                    | Description |
| ----------------------- | ----------- |
| --app-name | Name of the application where the service will be added.|
| --env-name | Name of the environment where the service will be added.|
| --git-repo-url | Optional.  Source Git repository URL.  It must be unique within GitOps.|
| --help | Shows help|
| --image-repo                          | Image repository of the form <registry>/<username>/<repository> or <project>/<app> which is used to push newly built images |
| --internal-registry-hostname          | Host-name for internal image registry e.g. docker-registry.default.svc.cluster.local:5000, used if you are pushing your images to the internal image registry |
| --pipelines-file | Optional.  Filepath to pipelines file.  Default is _pipelines.yaml_. |
| --sealed-secrets-ns string           | Optional. Namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator (default "cicd") |
| --sealed-secrets-svc string           | Optional. Name of the Sealed Secrets Services that encrypts secrets (default "sealedsecretcontroller-sealed-secrets") |
| --service-name | Name of the service to be added.  Service name must be unique within an environment. |
| --webhook-secret | Optional.  Optional. Provide a secret that we can use to authenticate incoming hooks from your Git hosting service.|

The following [directory layout](output) is generated.

```
.
├── apps
│   └── app-bus
│       ├── base
│       │   └── kustomization.yaml
│       ├── kustomization.yaml
│       ├── overlays
│       │   └── kustomization.yaml
│       └── services
│           └── bus
│               ├── base
│               │   └── kustomization.yaml
│               ├── kustomization.yaml
│               └── overlays
│                   └── kustomization.yaml
└── env
    ├── base
    │   ├── kustomization.yaml
    │   ├── new-env-environment.yaml
    │   └── new-env-rolebinding.yaml
    └── overlays
        └── kustomization.yaml
```

# KAM Service Command

## Service add

The `service add` sub-command adds a new service to an existing environment.

Services are logically grouped by applications.

This command will create an application for the new service if the target application does not exist.  It outputs resources YAML files, Kustomization files, and updated Manifest to filesystem.

**NOTE**: Service deployment resources are not generated.  They must be manually added to `environments/<env-new>/ services/<service-name>/base/config` and update the Kustomization file `environments/<env-new>/ services/<service-name>/base/kustomization.yaml`

```
dd a Service to an environment in GitOps

Usage:
  kam service add [flags]

Examples:
  Add a Service to an environment in GitOps
  kam service add

Flags:
      --app-name string                                Name of the application where the service will be added
      --env-name string                                Name of the environment where the service will be added
      --git-repo-url string                            GitOps repository e.g. https://github.com/organisation/repository
  -h, --help                                           help for add
      --image-repo string                              Image repository of the form <registry>/<username>/<repository> or <project>/<app> which is used to push newly built images
      --image-repo-internal-registry-hostname string   Host-name for internal image registry e.g. docker-registry.default.svc.cluster.local:5000, used if you are pushing your images to the internal image registry (default "image-registry.openshift-image-registry.svc:5000")
      --pipelines-folder string                        Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml (default ".")
      --sealed-secrets-ns string                       Namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator (default "kube-system")
      --sealed-secrets-svc string                      Name of the Sealed Secrets services that encrypts secrets (default "sealed-secrets-controller")
      --service-name string                            Name of the service to be added
      --webhook-secret string                          Source Git repository webhook secret (if not provided, it will be auto-generated)
```

```shell
$ kam service add
    --env-name
    --app-name
    --service-name
    [--git-repo-url]
    [--sealed-secrets-ns]
    [--webhook-secret]
    [--image-repo]
    [--image-repo-internal-registry-hostname]
    [--pipelines-folder]
```

The directory layout generated is shown below.
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

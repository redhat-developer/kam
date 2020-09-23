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

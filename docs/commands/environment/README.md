# Gitops Environment Command

## Environment add

The `environment add` sub-command creates a new environment in an existing GitOps setup.

It outputs resources YAML files, Kustomization files, and updated Manifest to filesystem.

```shell
$ gitops environment add
  --env-name 
  [--cluster]
  [--pipelines-file]
```

| Flag                    | Description |
| ----------------------- | ----------- |
| --cluster               | Deployment cluster (Default https://kubernetes.local.svc)|
| --env-name              | The name of environment to be added|
| --pipelines-file        | Optional.  Path to manifest file.  Default is pipelines.yaml. |
| --help                  | Show help|


The following [directory layout](output) is generated.

```
.
└── environments
    └── new-env
        └── env
            ├── base
            │   ├── kustomization.yaml
            │   └── new-env-environment.yaml
            └── overlays
                └── kustomization.yaml
```

# Gitops Environment Command

## Environment add

The `environment add` sub-command creates a new environment in an existing GitOps setup.

It outputs resources YAML files, Kustomization files, and updated Manifest to filesystem.

```shell
$ gitops environment add
  --env-name 
  [--cluster]
  [--pipelines-folder]
```

| Flag                    | Description |
| ----------------------- | ----------- |
| --cluster               | Deployment cluster (Default https://kubernetes.local.svc)|
| --env-name              | The name of environment to be added|
| --pipelines-folder      | Optional.  Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml |
| --help                  | Show help|


The directory layout generated is shown below.

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

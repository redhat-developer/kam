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

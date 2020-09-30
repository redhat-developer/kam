# KAM Environment Command

## Environment add

The `environment add` sub-command creates a new environment in an existing kam setup.

It outputs resources YAML files, Kustomization files, and updated Manifest to filesystem.

```
Add a new environment to the GitOps repository

Usage:
  kam environment add [flags]

Examples:
  # Add a new environment to GitOps
  kam environment add

Flags:
      --cluster string            Deployment cluster e.g. https://kubernetes.local.svc
      --env-name string           Name of the environment/namespace
  -h, --help                      help for add
      --pipelines-folder string   Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml (default ".")
```

```shell
$ kam environment add
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

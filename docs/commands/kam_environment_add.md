## kam environment add

Add a new environment

### Synopsis

Add a new environment to the GitOps repository

```
kam environment add [flags]
```

### Examples

```
  # Add a new environment to GitOps
  # Example: kam environment add --env-name new-env --pipelines-folder <path to GitOps folder>
  
  kam environment add
```

### Options

```
      --cluster string            Deployment cluster e.g. https://kubernetes.local.svc
      --env-name string           Name of the environment/namespace
  -h, --help                      help for add
      --pipelines-folder string   Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml (default ".")
```

### SEE ALSO

* [kam environment](kam_environment.md)	 - Manage an environment in GitOps


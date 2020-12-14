## kam service add

Add a new service

### Synopsis

Add a Service to an environment in GitOps

```
kam service add [flags]
```

### Examples

```
  Add a Service to an environment in GitOps
  kam service add
```

### Options

```
      --app-name string             Name of the application where the service will be added
      --env-name string             Name of the environment where the service will be added
      --git-repo-url string         GitOps repository e.g. https://github.com/organisation/repository
  -h, --help                        help for add
      --image-repo string           Image registry of the form <registry>/<username>/<image name> or <project>/<app> which is used to push newly built images
      --pipelines-folder string     Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml (default ".")
      --sealed-secrets-ns string    Namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator (default "cicd")
      --sealed-secrets-svc string   Name of the Sealed Secrets services that encrypts secrets (default "sealedsecretcontroller-sealed-secrets")
      --service-name string         Name of the service to be added
      --webhook-secret string       Source Git repository webhook secret (if not provided, it will be auto-generated)
```

### SEE ALSO

* [kam service](kam_service.md)	 - Manage services in an environment


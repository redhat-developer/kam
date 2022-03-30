## kam service add

Add a new service

### Synopsis

Add a Service to an environment in GitOps

```
kam service add [flags]
```

### Examples

```
  # Add a Service to an environment in GitOps
  # Example: kam service add --env-name new-env --app-name app-bus --service-name bus --git-repo-url http://github.com/<your organization>/bus.git --pipelines-folder <path to GitOps file>
  
  kam service add
```

### Options

```
      --app-name string           Name of the application where the service will be added
      --env-name string           Name of the environment where the service will be added
      --git-repo-url string       Service repository URL e.g. https://github.com/organisation/repository - only needed when you need to rebuild the source image for the environment
  -h, --help                      help for add
      --image-repo string         Image registry of the form <registry>/<username>/<image name> or <project>/<app> which is used to push newly built images
      --pipelines-folder string   Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml (default ".")
      --service-name string       Name of the service to be added
      --webhook-secret string     Source Git repository webhook secret (if not provided, it will be auto-generated)
```

### SEE ALSO

* [kam service](kam_service.md)	 - Manage services in an environment


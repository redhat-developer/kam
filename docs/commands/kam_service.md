## kam service

Manage services in an environment

### Synopsis

Manage services in a GitOps environment where service source repositories are synchronized

```
kam service [flags]
```

### Examples

```
kam service
add

  See sub-commands individually for more examples
```

### Options

```
      --app-name string           Name of the application where the service will be added
      --env-name string           Name of the environment where the service will be added
      --git-repo-url string       Service repository URL e.g. https://github.com/organisation/repository - only needed when you need to rebuild the source image for the environment
  -h, --help                      help for service
      --image-repo string         Image registry of the form <registry>/<username>/<image name> or <project>/<app> which is used to push newly built images
      --pipelines-folder string   Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml (default ".")
      --service-name string       Name of the service to be added
      --webhook-secret string     Source Git repository webhook secret (if not provided, it will be auto-generated)
```

### SEE ALSO

* [kam](kam.md)	 - kam
* [kam service add](kam_service_add.md)	 - Add a new service


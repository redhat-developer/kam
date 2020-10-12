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
      --app-name string                                Name of the application where the service will be added
      --env-name string                                Name of the environment where the service will be added
      --git-repo-url string                            GitOps repository e.g. https://github.com/organisation/repository
  -h, --help                                           help for service
      --image-repo string                              Image repository of the form <registry>/<username>/<repository> or <project>/<app> which is used to push newly built images
      --image-repo-internal-registry-hostname string   Host-name for internal image registry e.g. docker-registry.default.svc.cluster.local:5000, used if you are pushing your images to the internal image registry (default "image-registry.openshift-image-registry.svc:5000")
      --pipelines-folder string                        Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml (default ".")
      --sealed-secrets-ns string                       Namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator (default "cicd")
      --sealed-secrets-svc string                      Name of the Sealed Secrets services that encrypts secrets (default "sealedsecretcontroller-sealed-secrets")
      --service-name string                            Name of the service to be added
      --webhook-secret string                          Source Git repository webhook secret (if not provided, it will be auto-generated)
```

### SEE ALSO

* [kam](kam.md)	 - kam
* [kam service add](kam_service_add.md)	 - Add a new service


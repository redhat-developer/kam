## kam bootstrap

Bootstrap GitOps CI/CD with a starter configuration

### Synopsis

Bootstrap GitOps CI/CD Manifests

```
kam bootstrap [flags]
```

### Examples

```
  # Bootstrap OpenShift pipelines.
  kam bootstrap
```

### Options

```
      --SaveTokenKeyring                               Explicitely pass this flag to update the git-host-access-token in the keyring on your local file system
      --commit-status-tracker                          Enable or disable the commit-status-tracker which reports the success/failure of your pipelineruns to GitHub/GitLab (default true)
      --dockercfgjson string                           Filepath to config.json which authenticates the image push to the desired image registry  (default "~/.docker/config.json")
      --git-host-access-token string                   Used to authenticate repository clones, and commit-status notifications (if enabled). Access token is encrypted and stored on local file system by keyring, will be updated/reused.
      --gitops-repo-url string                         Provide the URL for your GitOps repository e.g. https://github.com/organisation/repository.git
      --gitops-webhook-secret string                   Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the GitOps repository. (if not provided, it will be auto-generated)
  -h, --help                                           help for bootstrap
      --image-repo string                              Image repository of the form <registry>/<username>/<repository> or <project>/<app> which is used to push newly built images
      --image-repo-internal-registry-hostname string   Host-name for internal image registry e.g. docker-registry.default.svc.cluster.local:5000, used if you are pushing your images to the internal image registry (default "image-registry.openshift-image-registry.svc:5000")
      --output string                                  Path to write GitOps resources (default ".")
      --overwrite                                      Overwrites previously existing GitOps configuration (if any)
  -p, --prefix string                                  Add a prefix to the environment names(Dev, stage,prod,cicd etc.) to distinguish and identify individual environments
      --private-repo-driver string                     If your Git repositories are on a custom domain, please indicate which driver to use github or gitlab
      --push-to-git                                    If true, automatically creates and populates the gitops-repo-url with the generated resources
      --sealed-secrets-ns string                       Namespace in which the Sealed Secrets operator is installed, automatically generated secrets are encrypted with this operator (default "cicd")
      --sealed-secrets-svc string                      Name of the Sealed Secrets Services that encrypts secrets (default "sealedsecretcontroller-sealed-secrets")
      --service-repo-url string                        Provide the URL for your Service repository e.g. https://github.com/organisation/service.git
      --service-webhook-secret string                  Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the Service repository. (if not provided, it will be auto-generated)
```

### SEE ALSO

* [kam](kam.md)	 - kam


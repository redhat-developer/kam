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
  kam bootstrap --service-repo-url https://github.com/<your organization>/taxi.git --gitops-repo-url https://github.com/<your organization>/gitops.git --image-repo quay.io/<username>/<image-repo> --dockercfgjson ~/Downloads/<username>-robot-auth.json --git-host-access-token <your git access token> --output <path to write GitOps resources> --push-to-git=true
  
  kam bootstrap
```

### Options

```
      --dockercfgjson string            Filepath to config.json which authenticates the image push to the desired image registry  (default "~/.docker/config.json")
      --git-host-access-token string    Used to authenticate repository clones. Access token is encrypted and stored on local file system by keyring, will be updated/reused.
      --gitops-repo-url string          Provide the URL for your GitOps repository e.g. https://github.com/organisation/repository.git
      --gitops-webhook-secret string    Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the GitOps repository. (if not provided, it will be auto-generated)
  -h, --help                            help for bootstrap
      --image-repo string               Image repository of the form <registry>/<username>/<repository> or <project>/<app> which is used to push newly built images
      --interactive                     If true, enable prompting for most options if not already specified on the command line
      --output string                   Path to write GitOps resources (default "./gitops")
      --overwrite                       Overwrites previously existing GitOps configuration (if any) on the local filesystem
  -p, --prefix string                   Add a prefix to the environment names(Dev, stage,prod,cicd etc.) to distinguish and identify individual environments
      --private-repo-driver string      If your Git repositories are on a custom domain, please indicate which driver to use github or gitlab
      --push-to-git                     If true, automatically creates and populates the gitops-repo-url with the generated resources
      --save-token-keyring              Explicitly pass this flag to update the git-host-access-token in the keyring on your local machine
      --service-repo-url string         Provide the URL for your Service repository e.g. https://github.com/organisation/service.git
      --service-webhook-secret string   Provide a secret that we can use to authenticate incoming hooks from your Git hosting service for the Service repository. (if not provided, it will be auto-generated)
```

### SEE ALSO

* [kam](kam.md)	 - kam


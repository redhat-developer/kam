## kam webhook create

Create a new webhook.

### Synopsis

Create a new Git repository webhook that triggers CI/CD pipeline runs.

```
kam webhook create [flags]
```

### Examples

```
  # Create a new Git repository webhook
  kam webhook create
```

### Options

```
      --access-token string       Access token to be used to create Git repository webhook
      --cicd                      Provide this flag if the target Git repository is a CI/CD configuration repository
      --env-name string           Provide environment name if the target Git repository is a service's source repository.
  -h, --help                      help for create
      --pipelines-folder string   Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml (default ".")
      --service-name string       Provide service name if the target Git repository is a service's source repository.
```

### SEE ALSO

* [kam webhook](kam_webhook.md)	 - Manage Git repository webhooks


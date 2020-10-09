## kam webhook delete

Delete webhooks.

### Synopsis

Delete all Git repository webhooks that trigger event to CI/CD Pipeline Event Listeners.

```
kam webhook delete [flags]
```

### Examples

```
  # Delete a Git repository webhook
  kam webhook delete
```

### Options

```
      --access-token string       Access token to be used to create Git repository webhook
      --cicd                      Provide this flag if the target Git repository is a CI/CD configuration repository
      --env-name string           Provide environment name if the target Git repository is a service's source repository.
  -h, --help                      help for delete
      --pipelines-folder string   Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml (default ".")
      --service-name string       Provide service name if the target Git repository is a service's source repository.
```

### SEE ALSO

* [kam webhook](kam_webhook.md)	 - Manage Git repository webhooks


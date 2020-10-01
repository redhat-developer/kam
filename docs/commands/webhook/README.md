# KAM Webhook Command

* [webhook create](#Webhook-create)
* [webhook delete](#Webhook-delete)
* [webhook list](#Webhook-list)

## Webhook create

The `webhook create` sub-command creates a webhook on target Git repository using secret and EventListener address retrieved from cluster.

If a webhook (with the same EventListener address URL) already exists, a webhook will not be created.

Otherwise, a webhook will be created and the ID of the new webhook is written to standard output.

```
Create a new Git repository webhook that triggers CI/CD pipeline runs.

Usage:
  kam webhook create [flags]

Examples:
  # Create a new Git repository webhook
  kam webhook create

Flags:
      --access-token string       Access token to be used to create Git repository webhook
      --cicd                      Provide this flag if the target Git repository is a CI/CD configuration repository
      --env-name string           Provide environment name if the target Git repository is a service's source repository.
  -h, --help                      help for create
      --pipelines-folder string   Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml (default ".")
      --service-name string       Provide service name if the target Git repository is a service's source repository.
```

## Webhook delete

The `webhook delete` sub-command deletes all webhooks from Git repository that contains the target EventListener address.

The EventListener address is retrieved from cluster based on the options passed to the command. The IDs of the deleted webhooks will be written to standard output.

```
Delete all Git repository webhooks that trigger event to CI/CD Pipeline Event Listeners.

Usage:
  kam webhook delete [flags]

Examples:
  # Delete a Git repository webhook
  kam webhook delete

Flags:
      --access-token string       Access token to be used to create Git repository webhook
      --cicd                      Provide this flag if the target Git repository is a CI/CD configuration repository
      --env-name string           Provide environment name if the target Git repository is a service's source repository.
  -h, --help                      help for delete
      --pipelines-folder string   Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml (default ".")
      --service-name string       Provide service name if the target Git repository is a service's source repository.
```      

## Webhook list

The `webhook list` sub-command displays webhook IDs from the Git repository that contain the target EventListener address.

The EventListener address is retrieved from cluster based on the options passed to the command. The IDs of the found webhooks will be written to standard output.

```
List existing Git repository webhook IDs of the target repository and listener.

Usage:
  kam webhook list [flags]

Examples:
  # List Git repository webhook IDs
  kam webhook list

Flags:
      --access-token string       Access token to be used to create Git repository webhook
      --cicd                      Provide this flag if the target Git repository is a CI/CD configuration repository
      --env-name string           Provide environment name if the target Git repository is a service's source repository.
  -h, --help                      help for list
      --pipelines-folder string   Folder path to retrieve manifest, eg. /test where manifest exists at /test/pipelines.yaml (default ".")
      --service-name string       Provide service name if the target Git repository is a service's source repository.
```

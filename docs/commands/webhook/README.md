# Gitops Webhook Command

* [webhook create](#Webhook-create)
* [webhook delete](#Webhook-delete)
* [webhook list](#Webhook-list)

## Webhook create

The `webhook create` sub-command creates a webhook on target Git repository using secret and EventListener address retrieved from cluster.

If a webhook (with the same EventListener address URL) already exists, a webhook will not be created.

Otherwise, a webhook will be created and the ID of the new webhook is written to standard output.

```shell
$ gitops webhook create 
    --access-token 
    [--cicd] | [--env-name --service-name]
    [--pipelines-folder]
```

## Webhook delete

The `webhook delete` sub-command deletes all webhooks from Git repository that contains the target EventListener address.

The EventListener address is retrieved from cluster based on the options passed to the command. The IDs of the deleted webhooks will be written to standard output.

```shell
$ gitops webhook delete
    --access-token
    [--cicd] | [--env-name --service-name]
    [--pipelines-folder ]
```

## Webhook list

The `webhook list` sub-command displays webhook IDs from the Git repository that contain the target EventListener address.

The EventListener address is retrieved from cluster based on the options passed to the command. The IDs of the found webhooks will be written to standard output.

```shell
$ gitops webhook list
    --access-token
    [--cicd] | [--env-name --service-name]
    [--pipelines-folder ]
```
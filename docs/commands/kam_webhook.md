## kam webhook

Manage Git repository webhooks

### Synopsis

Add/Delete/list Git repository webhooks that trigger CI/CD pipeline runs.

```
kam webhook [flags]
```

* When a token is provided in command flag `--access-token`, the token will be stored securely in keyring. The hostname of the URL (e.g. github.com) will be used to stored the token by keyring. The provided token will be used.

* When a token is not provided in command flag `--access-token`, the token is retrieved by keyring using the hostname of the URL. If no token is not found, a token is retrieved from env var. What is the format of the env var? If no token is not found in the env var, the command will fail.

* When a token is not provided in command flag and the token is not present in the keyring, the webhook cmd will look for the access token in an environment variable with the syntax `HOSTNAME_TOKEN` (e.g. GITHUB_COM_TOKEN).

### Examples

```
kam webhook
create
delete
list

  See sub-commands individually for more examples
```

### Options

```
  -h, --help   help for webhook
```

### SEE ALSO

* [kam](kam.md)	 - kam
* [kam webhook create](kam_webhook_create.md)	 - Create a new webhook.
* [kam webhook delete](kam_webhook_delete.md)	 - Delete webhooks.
* [kam webhook list](kam_webhook_list.md)	 - List existing webhook Ids.


## kam webhook

Manage Git repository webhooks

### Synopsis

Add/Delete/list Git repository webhooks that trigger CI/CD pipeline runs.

```
kam webhook [flags]
```
* When a token is provided in command flag `--access-token`, the token will be stored securely in keyring. The hostname of the URL (e.g. github.com) will be used to store the token by keyring. The provided token will be used.

* When a token is not provided in command flag `--access-token`, the token is retrieved by keyring using the `hostname of the URL` as the username and service name `kam`. If no token is not found in the keyring, a token is retrieved from environment variable as described in the step below.

* When a token is not provided in command flag and the token is not present in the keyring, the webhook cmd will look for the access token in an environment variable with the syntax HOSTNAME_TOKEN (e.g. GITHUB_COM_TOKEN). The environment variable name is assigned as follows, the hostname (e.g. github.com) is extracted from the value passed to the repository URL (e.g. https://github.com/username/repo.git),  where the `.` in the hostname is replaced by `_` and concatenated with `_TOKEN`. Considering the previous examples, the environment varaible name will be `GITHUB_COM_TOKEN`.

Assuming the token is not passed in the command, if the token is not found in the keyring or the environment variable with the specified name. The command will fail.

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


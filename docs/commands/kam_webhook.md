## kam webhook

Manage Git repository webhooks

### Synopsis

Add/Delete/list Git repository webhooks that trigger CI/CD pipeline runs.

```
kam webhook [flags]
```
#### Important note:

The webhoook command first searches for the personal-access-token(a.k.a access-token) in the key-ring on your local file system. The secret is stored at the time of bootsrap with the following credentials service name ```kam``` and username being the hostname of the gitops-repo. Eg. ```github.com``` being the hostname for gitops repo  ```URL: https://github.com/user/abc.git``` . If the git-host-access-token is not found in the keyring, it looks for the access-token in an environment variable with the name ```[REPO-NAME]GITTOKEN```. Eg. For the aforementioned ```URL: https://github.com/user/abc.git```, the name of the environment variable that is expected to possess the access-token is ```ABCGITTOKEN```. Kindly note, this behavior can be overriden by passing the ```access-token``` flag.


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


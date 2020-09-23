## Quay Credentials to push built image to Quay.io registry

Some of the Tasks in this Tutorial involve pushing images to Quay image registry.   (The image is to be built by CI Pipeline.)   Before we can start creating Kubernetes resources, we need to obtain credentials for your Quay user account.

 * Create `taxi` Quay repos. Login to your Quay.io account and create a repository `taxi`

 ![Screenshot](img/create-taxi-in-quay.png)

 * Login to your Quay.io account that you can generate read/write credentials for.  In user's name pulldown menu, goto Account Settings -> Robot Account (button on the left).   Create a robot account for yourself.  Click your robot account link.

 ![Screenshot](img/quay-create-robot-account.png)

 * Select `Edit Repository Permissions` menu item

 ![Screenshot](img/edit-token-permission.png)

 * Grant `write` permission to repository `taxi`

 ![Screenshot](img/grant-write-permission.png)

 * Download Docker Configuration file to `<Quay user>-robot-auth.json`

 ![Screenshot](img/quay-download-docker-config.png)

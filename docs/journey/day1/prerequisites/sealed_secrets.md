## Suggested approach

First, we need a cicd namespace:

```shell
$ oc create namespace cicd
```

Install the Sealed Secrets from the Operator Hub.

![Screenshot](img/sealed-secrets-operator.png)

Then create a "SealedSecretController" instance in the "cicd" namespace.

![Screenshot](img/sealed-secrets-controller-in-cicd.png)

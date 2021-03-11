# Suggested approach

First, create a `cicd` namespace:

```shell
$ oc create namespace cicd
```

Install the Sealed Secrets Operator from the Operator Hub in the `cicd` namespace.

![Screenshot](img/ss-1.png)

![Screenshot](img/ss-2.png)

![Screenshot](img/ss-3.png)

![Screenshot](img/ss-4.png)



Then create a `SealedSecretController` instance in the `cicd` namespace.

![Screenshot](img/ss-5.png)

![Screenshot](img/ss-6.png)

![Screenshot](img/ss-7.png)

![Screenshot](img/ss-8.png)





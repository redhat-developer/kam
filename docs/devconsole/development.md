## Contribution and developing locally

1. To set up a local development environment for OpenShift Console, first clone the [console repo](https://github.com/openshift/console) and `cd` into that directory. 

2. Login to an OpenShift cluster with `oc login`  and then run the following steps:

```
$ source ./contrib/oc-environment.sh
$ ./build.sh
$ ./bin/bridge
```

3. For allowing our local UI development environment to connect to the gitops backend, we need to expose our backend using a route. Create a cluster route file `cluster-route.yaml` and add the following route to it:
```
kind: Route
apiVersion: route.openshift.io/v1
metadata:
  name: cluster
  namespace: openshift-gitops  
spec: 
  to: 
    kind: Service
    name: cluster   
  port:
    targetPort: 8080
  tls:
    termination: reencrypt
    insecureEdgeTerminationPolicy: Allow
```   

4. Run `oc apply -f cluster-route.yaml`, this will create a route that connects to the backend.

5. Now you may continue with [Day 1](../journey/day1/README.md) and [Day 2](../journey/day2/README.md) Kam Bootstrapping operations to add gitops manifests and populate the UI. 
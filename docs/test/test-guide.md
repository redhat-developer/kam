# Test Guide

## Setting up test environment

The minimum version of Go required is in the [https://github.com/redhat-developer/kam/blob/master/go.mod#L3](go.mod) file.

### Tests

#### Unit tests:

Unit test does not require any cluster configuration, run `make test` to validate unit tests.

##### Prerequisites for OpenShift cluster:

* A `crc` environment for 4.5+ local cluster:
Follow [https://github.com/code-ready/crc#documentation](crc) installation guide.
* Or a 4.5+ cluster hosted remotely

NOTE: Make sure that `kam` and `oc` binaries are in `$PATH`. Use the cloned kam directory to launch tests on `4.5+` clusters. `4.5+` cluster needs to be configured before launching the tests against it. The files `kubeadmin-password` and `kubeconfig` which contain cluster login details should be present in the `auth` directory and it should reside in the same directory as `Makefile`. If it is not present in the auth directory, please create it. Then run `make prepare-test-cluster` to configure the `4.5+` cluster. `make prepare-test-cluster` comprises installation of sealed secrets, openshift pipelines, argocd operator and create sealed secrets instance.

#### E2e tests:
//TODO

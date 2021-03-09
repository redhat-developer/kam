# Test Guide

## Setting up test environment

The minimum version of Go required is in the [https://github.com/redhat-developer/kam/blob/master/go.mod#L3](go.mod) file.

### Tests

#### Unit tests:

Unit test does not require any cluster configuration, run `make test` to validate unit tests. Unit tests run on OpenShift CI before the code is merged. Unit test coverage is then reported to Codecov which provides coverage information on the pull requests.

To run tests during the development cycle:
```
$ make test
```
To run specific tests, use one of the following methods:

* Run all tests on a single package.
    ```
    # Eg: go test -v ./pkg/cmd/environment
    $ go test -v <relative path of package>
    ```

* Run a single test on a single package.
    ```
    $ go test -v <relative path of package> -run <Testcase Name>
    ```
    
* Run tests that match a pattern.
    ```
    $ go test -v <relative path of package> -run "Test<Regex pattern to match tests>"
    ```
For more information about test options, run the `go test --help` command and review the documentation.

#### Prerequisites for OpenShift cluster:

* A `crc` environment for 4.5+ local cluster:
Follow [https://github.com/code-ready/crc#documentation](crc) installation guide.
* Or a 4.5+ cluster hosted remotely

NOTE: Make sure that `kam`, `oc`, `gh` and `argocd` binaries are in `$PATH`. Use the cloned kam directory to launch tests on `4.5+` clusters. `4.5+` cluster needs to be configured before launching the tests against it. The files `kubeadmin-password` and `kubeconfig` which contain cluster login details should be present in the `auth` directory and it should reside in the same directory as `Makefile`. If it is not present in the auth directory, please create it. Then run `make prepare-test-cluster` to configure the `4.5+` cluster. `make prepare-test-cluster` comprises installation of sealed secrets and OpenShift GitOps operator and create sealed secrets instance.

#### E2e tests:
E2e(end to end) tests utilize [godog](https://github.com/cucumber/godog) and an external library package [clicumber](https://github.com/code-ready/clicumber) which define sets of generic gherkin test steps.

Clicumber allows running commands in a persistent shell instance (bash, tcsh, zsh, Command Prompt, or PowerShell), assert its outputs (standard output, standard error, or exit code), check configuration files, and so on.

Kam test feature files are located in `tests/e2e` directory and can be called using `make e2e`.

#### How to write the test feature files

Before writing KAM specific steps make sure that the same step is not part of [clicumber](https://github.com/code-ready/clicumber/blob/master/testsuite/testsuite.go) generic steps.

In kam suite we can add a full test step for our reference. First wite the test step skeleton and then write its backend implementation.

For example:

In the kamsuite.go, write the test step skeleton
```
s.Step(`^directory "([^"]*)" should exist$`, DirectoryShouldExist)
```

This defines a step which matches regular expression `^directory "([^"]*)" should exist$`. If matched, the capturing group is passed as parameters to the function `DirectoryShouldExist(dirName string)`, which implements the actual behaviour of the step that is:
```
func DirectoryShouldExist(dirName string) error {
	if _, err := os.Stat(dirName); os.IsExist(err) {
		return nil
	}

	return fmt.Errorf("directory %s exists", dirName)
}
```
To use the step in the feature file, you need to make sure it matches with the regular expression and prepend one of the Gherkin keywords: `Given, When, Then, And or But`, for example:
```
And directory "bootstrapresources" should exist
```
NOTE: See the [Gherkin Reference](https://cucumber.io/docs/gherkin/reference/) for more general information about the structure of Gherkin, its features, scenarios, and steps.

#### Run E2E test locally

To run the e2e test locally, user need to export the environment variables SERVICE_REPO_URL, GITOPS_REPO_URL, IMAGE_REPO, DOCKERCONFIGJSON_PATH and GITHUB_TOKEN corresponding to its flag --service-repo-url, --gitops-repo-url, --image-repo, --dockercfgjson and --git-host-access-token respectively.

For example:
```
$ export SERVICE_REPO_URL=<Provide the URL for your Service repository>
$ export GITOPS_REPO_URL=<Provide the URL for your GitOps repository>
$ export IMAGE_REPO=<Image repository which is used to push newly built images>
$ export DOCKERCONFIGJSON_PATH=<Filepath to config.json which authenticates the image push to the desired image registry>
$ export GITHUB_TOKEN=<Used to authenticate repository clones, and commit-status notifications (if enabled)>
```

Then run the command `make e2e`.

#### Using the GODOG_OPTS Parameter

The `GODOG_OPTS` parameter specifies additional arguments for the Godog runner. The following options are available:

* Tags

    Use tags to ensure that scenarios and features containing at least one of the selected tags are executed. To select particular feature, you can use its name as a tag. For example, the basic.feature contains @automated tag through which it can be selected and run with the following command: `make e2e GODOG_OPTS=--tags=basic`. There are also a few special tags used to indicate specific subsets of e2e tests. These are the following:

* Paths

    Use paths to define paths to different feature files or folders containing feature files. This can be used to run feature files outside of the test/e2e/features folder.

* Format

    Use format to change the format of Godog’s output. For example, you can set the format to progress instead of the default pretty.

* Stop-on-failure

    Set stop-on-failure to true to stop e2e tests on failure.

* No-colors

    Set no-colors to true to disable ANSI colorization of Godog’s output.

* Definitions

    Set definitions to true to print all available step definitions.

Note: Passing any value via `GODOG_OPTS` overrides the default tag definition on each e2e target. Thus in this case `--tags` must be specified manually, otherwise all features will be run.

For example, to run e2e tests on two specific feature files using only the @automated tags and without ANSI colors, the following command can be used:
```
$ make e2e GODOG_OPTS="-paths ~/tests/custom.feature,~/my.feature -tags basic -no-colors true"
```
NOTE: Multiple values for a `GODOG_OPTS` option must be separated by a comma without whitespace. For example, -tags basic,manual will be parsed properly by make, whereas -tags basic, manual will result in only @basic being used.

#### Viewing Results

The e2e test logs its progress directly into a console. This information is often enough to find and debug the reason for a failure.

However, for cases which require further investigation, the e2e test also logs more detailed progress into a log file. This file is located at $GOPATH/github.com/redhat-developer/kam/out/test-results/integration_YYYY-MM-DD_HH:MM:SS.log.


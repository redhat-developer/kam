# GitOps Application Manager

The GitOps Application Manager provides command line interface to bootstrap GitOps and perform other operations.  See why and how [GitOps for Application Delivery](./docs/README.md) can help your business streamline your application delivery.

## CLI Reference

[Command Line Reference](./docs/commands/README.md)

## Getting Started

### GitOps Day 1 and Day 2 operations
* [Day 1 Operations](docs/journey/day1): Install the prerequisites and setup your GitOps pipeline.
* [Day 2 Operations](docs/journey/day2): Continue adding more applications.
  


Please visit the CLI [user documentation](./docs/README.md) to try out the CLI. For more information regarding how kam CLI is used, please refer to this [blog](https://developers.redhat.com/articles/2021/07/21/bootstrap-gitops-red-hat-openshift-pipelines-and-kam-cli).

### OpenShift Console

Please visit the [OpenShift Console Documentation](./docs/devconsole) to visualize Environments on your Console's Developer Perspective.

### FAQs

[GitOps Frequently Asked Questions](./docs/FAQ/GitopsFAQ.md)

## How to Contribute

Building `kam` requires Go 1.14+

To contribute to `KAM CLI`, follow these steps:

1. Fork this repository.
2. Create a branch: `git checkout -b <branch_name>`.
3. Make your changes and commit them: `git commit -m '<commit_message>'`
4. Push to the original branch: `git push origin <project_name>/<location>`
5. Create the pull request.

Alternatively see the GitHub documentation on [creating a pull request](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request).

### How to Build

```shell
$ make bin
```

The resulting binary will be found in `./bin/kam`

Alternatively this is a Standard Go project layout project, so you can build with:

```shell
$ go build ./cmd/kam
```

## Contact

Please open a Github Issue or reach out to the team at [team-gitops@redhat.com](mailto:team-gitops@redhat.com)

## License

This project uses the following license: [Apache 2.0](./LICENSE).

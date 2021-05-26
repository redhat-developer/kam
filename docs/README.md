# GitOps for Application Delivery

## Why

As a software organization, I would like to:

* Audit all changes made to pipelines, infrastructure, and application
  configuration.
* Roll forward/back to desired state in case of issues.
* Consistently configure all environments.
* Reduce manual effort by automating application and environment setup and remediation.
* Have an easy way to manage application and infrastructure state across clusters/environments.

## What

* GitOps is a natural evolution of DevOps and Infrastructure-as-Code.
* GitOps is when the infrastructure and/or application state is fully represented by the contents of a git repository. Any changes to the git repository are reflected in the corresponding state of the associated infrastructure and applications through automation.

## Principles

* Git is the source of truth.
* Separate application source code (Java/Go) from deployment manifests i.e the application source code and the GitOps configuration reside in separate git repositories.
* Deployment manifests are standard Kubernetes (k8s) manifests i.e Kubernetes manifests in the GitOps repository can be simply applied with nothing more than a `oc apply`.
* Kustomize for defining the differences between environments i.e reusable parameters with extra resources described using kustomization.yaml.

## Installation

Please download the binary for your Operating System from the latest release:

[https://github.com/redhat-developer/kam/releases/latest](https://github.com/redhat-developer/kam/releases/latest)

### Steps

The following steps will guide you to setup GitOps for application delivery:

1. [GitOps Day 1 Operations](./journey/day1): Install the prerequisites and setup your GitOps Pipeline.
2. [GitOps Day 2 Operations](./journey/day2): Continue adding more applications.

### Reference

The command reference to generate and manage the above can be found [here](./commands).

This project provides the tooling to generate the recommended directories and manifests as per the [Pipelines Model](./model).

### Sample

A sample GitOps repository can be found [here](https://github.com/redhat-developer/gitops-repo-example).

## Business Scenario

A company called *Pet Clinic Supplies Inc* has a big vision, but limited investment, they decide to start developing code for their online portal that currently sells supplies for pet clinics.

They begin by developing code with five engineers who contribute to their version control system and perform testing, quality, and security checks and deploy to Kubernetes clusters themselves. As the release frequency increased, they realized they were spending more time managing deployments rather than developing features. They wanted to streamline the deployment process along with development by auditing the changes to the configurations (application, infrastructure, pipelines, etc) in git.  

They now need a quick fix, on a quick Google search they decide to incorporate GitOps, as a [GitOps Day 1 Operations](./journey/day1) they want to quickly get started and get their deployments up and running. This is when they use our GitOps tools `kam bootstrap` command to quickly set up their _dev_ and _stage_ environments, along with pipelines for seamless CI/CD with the bootstrapped environments, incorporating the [Pipelines Model](./model).

Now that we've given the team its ability and power to use version control to monitor progress on its different environments, the team decides to make modifications to its current bootstrapped GitOps configuration, they want to incorporate their own online wallet for users to shop and spend more easily. This requires the use of an additional security environment and that has specific microservices to test traffic or test vulnerabilities only particular to this environment, the team has to look no further as they can leverage the use of our [GitOps Day 2 Operations](./journey/day2), simply running the `kam add environment` command to add their security environment, and choose the services they wish to add to the environment with a simple `kam service add` command.

The bootstrapped environment sets up the necessary pipelines using OpenShift Pipelines required for developers/teams to perform CI with their respective environments, they can then check the progress of their deployment as CD for the applications is performed using [Argo CD](https://argoproj.github.io/argo-cd/) constantly synchronizing the desired state of the application to the cluster. 

By adopting GitOps, the team can now follow the well known git workflow and outsource the deployment process. Deploying to different environments across multiple clusters is now as simple as merging a pull request. Every change to the configuration is well audited and safe, since “Git is the single source of truth”.

Now's your chance to ramp up your organization using our modernized approach for end-to-end application delivery.

A sample GitOps configuration can be seen [here](https://github.com/redhat-developer/gitops-repo-example), all achieved with a list of commands listed in [kam Command Reference](./commands), why wait?

**NOTE**: The sample repository cannot be applied to a cluster, the secrets are encrypted and the key is not available.

Lets get started.

## Support for Git hosting services

GitHub and GitLab are supported. However, only one Git driver is supported during bootstrap.

The Git driver is determined by the GitOps Repository URL used during bootstrapping/initialization.

For example, if a GitHub repository URL is specified during bootstrapping, all service/application Git repositories must be GitHub repositories.

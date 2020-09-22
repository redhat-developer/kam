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

[https://github.com/redhat-developer/gitops-cli/releases/latest](https://github.com/redhat-developer/gitops-cli/releases/latest)

### Steps

The following steps will guide you to setup GitOps for application delivery:

1. [GitOps Day 1 Operations](./journey/day1): Install the prerequisites and setup your GitOps Pipeline.
2. [GitOps Day 2 Operations](./journey/day2): Continue adding more applications.

### Reference

The command reference to generate and manage the above can be found [here](./commands).

This project provides the tooling to generate the recommended directories and manifests as per the [Pipelines Model](./model).

### Sample

A sample GitOps repository can be found [here](https://github.com/rhd-gitops-example/gitops).

## Business Scenario

A company called *Pet Clinic Supplies Inc* has a big vision, but limited investment, they decide to start developing code for their online portal that currently sells supplies for pet clinics.

They begin by developing code using legacy systems with five engineers who contribute to their own version control system/database and perform testing, quality and security checks themselves, all the code and business decisions can be reviewed within the team. Soon they start to notice a sudden surge in the business and to broaden their business they decide to incorporate more commodities (i.e hardware). To make this happen they now need more developers and realize that their current methods are no longer working well for them.

They can't really tell who makes or breaks their code, there's a loss of accountability.

They now need a quick fix, on a quick Google search they decide to incorporate GitOps, as a [GitOps Day 1 Operations](./journey/day1) they want to quickly get started to get their teams up and running.

This is when they use our GitOps tools `gitops bootstrap` command to quickly set up their _dev_ and _stage_ environments, along with pipelines for seamless CI/CD with the bootstrapped environments, incorporating the [Pipelines Model](./model).

Now the management is happy as they can keep track of the different teams within the organization that are making changes to their code-base and can hold the teams accountable for their actions.

Now that we've given the team its ability and power to use version control to monitor progress on its different teams/environments, the engineering management decides to make modifications to its current bootstrapped GitOps configuration, they want to incorporate their own online wallet for users to shop and spend more easily. This requires the use of an additional security environment and that has specific microservices to test traffic or test vulnerabilities only particular to this environment, the management has to look no further as they can leverage the use of our [GitOps Day 2 Operations](./journey/day2), simply running the `gitops add environment` command to add their security environment, and choose the services they wish to add to the environment with a simple `gitops service add` command.

They now need a quick fix, on a quick Google search they decide to incorporate GitOps, as a [GitOps Day 1 Operations](./journey/day1) they want to quickly get started to get their teams up and running. This is when they use our GitOps tools `gitops bootstrap` command to quickly set up their dev, stage and prod environment along with the config repos that create a pipeline for seamless CI/CD with the bootstrapped environments incorporating the [Pipelines Model (aka manifest model)](./model). Now the management is happy as they can keep track of the different teams within the organization that are making changes to their code base and can hold the teams accountable for their actions.

Now that we’ve given the team its ability and power to use version control to monitor progress on its different teams/environments , the engineering management decides to make modifications to its current bootstrapped GitOps configuration, they want to incorporate their own online wallet for users to shop and spend more easily. But this requires the use of an additional security environment and that has specific micro-services to test traffic or test vulnerability only particular to this environment, the management has to look no further as they can leverage the use of our [GitOps Day 2 Operations](./journey/day2), simply running the `gitops add environment` command to add their security environment, and choose the services they wish to add to the environment with a simple `gitops service add` command.

The bootstrapped environment sets up the necessary pipelines using OpenShift Pipelines required for developers/teams to perform CI with their respective environments, they can then check the progress of their development as CD for the applications is performed using [ArgoCD](https://argoproj.github.io/argo-cd/) constantly synchronising the desired state of the application to the cluster.

Users can easily customize the pipelines to their own needs. A simple push to the GitOps repo can alter the configuration of the GitOps repo and hence the style by which you want to run your organization, by triggering a CI pipeline, further a merge to the trunk branch triggers a deployment that further changes the state of your application on the cluster. Giving teams to use the power of OpenShift/Kubernetes to run powerful applications.

Now we have an a team that set out as a start-up but went on to be a full-fledged organization with multiple verticals and businesses set up in different geographical locations leveraging the use of OpenShift and GitOps with "Git as the source of truth".

Now's your chance to ramp up your organization using our modernized approach for end-to-end application delivery.

A simple GitOps configuration can be seen [here](https://github.com/rhd-gitops-example/gitops), all achieved with a list of commands listed in [gitops Command Reference](./commands), why wait?

Lets get started.

## Support for Git hosting services

GitHub and GitLab are supported. However, only one Git driver is supported during bootstrap.

The Git driver is determined by the GitOps Repository URL used during bootstrapping/initialization.

For example, if a GitHub repository URL is specified during bootstrapping, all service/application Git repositories must be GitHub repositories.

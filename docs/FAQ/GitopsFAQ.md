# Frequently Asked Questions

## What is GitOps?
_GitOps is a way to do operations, by using Git as a single source of truth, and updating the state of the operating configuration automatically, based on a Git repository_.

## How does GitOps differ from Infrastructure as Code?
_GitOps builds on top of Infrastructure as Code, providing application level concerns, as well as an operations model_.

## Can I use a CI server to orchestrate convergence in the cluster?
_You could apply updates to the cluster from the CI server, but it won’t continuously deploy the changes to the cluster, which means that drift won’t be detected and corrected._

## Should I abandon my CI tool?
_No, you'll want  CI to validate the changes that GitOps is applying._

## Why choose Git and not a Configuration Database instead? / Why is git the source of truth?
_Git has strong auditability, and it fits naturally into a developer's flow._

## How do you keep my tokens secret in the Git repository?
_We are going with Sealed Secrets because of it's low-maintenance, and because it requires little investment to get going, you need to consider that anything you put into Git might get leaked at some point, so if you’re keeping secrets in there, they might be made publicly available._

## How do I get started?
_Add some resources to a directory, and git commit and push, then ask Argo CD to deploy the repository, change your resource, git commit and push, and the change should be deployed automatically._

## How are OpenShift pipelines used?
_They are used in the default setup to drive the CI from pushes to your application code repository_.

## How is GitOps different from DevOps?
_GitOps is a subset of DevOps, specifically focussed on deploying the application (and infrastructure) through a Git flow-like process._

## How could small teams benefit from GitOps?
_GitOps is about speeding up application feedback loops, with more automation, it frees up developers to work on the product features that customers love._

## I have a non-globally trusted certificate in front of my private GitHub/GitLab installation, how do I get it to work?
You'll need to reconfigure the automatically generated pipeline resources. Append `sslVerify` parameter in the following files:

1. `config/cicd/base/06-templates/ci-dryrun-from-push-template.yaml`

```yaml
      pipelineRef:
        name: ci-dryrun-from-push-pipeline
      resources:
      - name: source-repo
        resourceSpec:
          params:
          - name: revision
            value: $(tt.params.io.openshift.build.commit.id)
          - name: url
            value: $(tt.params.gitrepositoryurl)
          - name: sslVerify
            value: "false"
          type: git
```

2. `config/cicd/base/04-pipelines/app-ci-pipeline.yaml`

```yaml
      tasks:
      - name: clone-source
        params:
        - name: url
          value: $(params.GIT_REPO)
        - name: revision
          value: $(params.GIT_REF)
        - name: sslVerify
          value: "false"
        taskRef:
          kind: ClusterTask
          name: git-clone
        workspaces:
        - name: output
          workspace: shared-data
```

```yaml
          - name: sslVerify
            value: "false"
```

This additional parameter configures the TLS to be insecure, i.e. it will not do _any_ validation of the TLS certificate that the server presents, so yes, the data is encrypted, but you don't know who you are sending it to.

You will also need to configure Argo CD to fetch your data insecurely.

```
$ argocd repo add https://gitlab.example.com/my-org/my-gitops-repo.git --username git --password <auth token> --insecure-skip-server-verification
```

## The secrets in my Git repository are encrypted, how do I backup the key?

https://github.com/bitnami-labs/sealed-secrets#how-can-i-do-a-backup-of-my-sealedsecrets



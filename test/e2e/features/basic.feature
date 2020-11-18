Feature: Basic test
    Checks whether KAM top-level commands behave correctly.

    Scenario: KAM version
        When executing "kam version" succeeds
        Then stderr should be empty
        And stdout should contain "kam version"

    Scenario: Execute KAM bootstrap command without --push-to-git=true flag
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GITHUB_TOKEN --output bootstrapresources" succeeds
        Then stderr should be empty

        When executing "cd bootstrapresources" succeeds
        When executing "git init ." succeeds
        When executing "git add pipelines.yaml config environments" succeeds
        When executing "git commit -m 'Bootstrapped commit'" succeeds
        When executing "git branch -m main" succeeds
        When executing "git remote add origin $GITOPS_REPO_URL" succeeds
        When executing "git push -u $GITOPS_REPO_URL main" succeeds

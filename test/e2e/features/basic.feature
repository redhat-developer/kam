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
        Then executing "git init ." succeeds
        Then executing "git add ." succeeds
        Then executing "git commit -m 'Initial commit.'" succeeds
        Then executing "git branch -m main" succeeds
        Then executing "git remote add origin $GITOPS_REPO_URL" succeeds
        Then executing "git push -u $GITOPS_REPO_URL main" succeeds

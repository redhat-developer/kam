Feature: Basic test
    Checks whether KAM top-level commands behave correctly.

    Scenario: KAM version
        When executing "kam version" succeeds
        Then stderr should be empty
        And stdout should contain "kam version"

    Scenario: KAM bootstrap command
        When executing 'echo -e "Host github.com\n\tStrictHostKeyChecking no\n" >> ~/.ssh/config' succeeds
        Then executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GITHUB_TOKEN --output bootstrapresources --push-to-git=true" succeeds
        Then stderr should be empty

Feature: Basic test
    Checks whether KAM top-level commands behave correctly.

    Scenario: KAM version
        When executing "kam version" succeeds
        Then stderr should be empty
        And stdout should contain "kam version"

    Scenario: KAM bootstrap
        When executing "kam bootstrap \
        --service-repo-url $SERVICE_REPO_URL \
        --gitops-repo-url $GITOPS_REPO_URL \
        --image-repo $GITOPS_REPO_URL \
        --dockercfgjson $DOCKERCONFIGJSON_PATH \
        --git-host-access-token $GIT_HOST_ACCESS_TOKEN \
        --output bootstrapresources \
        --push-to-git=true" succeeds
        Then stderr should be empty

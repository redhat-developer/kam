Feature: Basic test
    Checks whether KAM top-level commands behave correctly.

    Scenario: KAM version
        When executing "kam version" succeeds
        Then stderr should be empty
        And stdout should contain "kam version"

    Scenario: Execute KAM bootstrap command without --push-to-git=true flag
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GITHUB_TOKEN --output bootstrapresources" succeeds
        Then stderr should be empty
    
    Scenario: KAM bootstrap command with --push-to-git=true and --overwrite flag
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GITHUB_TOKEN --output bootstrapresources --push-to-git=true --overwrite" succeeds
        Then stderr should be empty

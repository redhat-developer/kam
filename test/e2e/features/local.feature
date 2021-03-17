Feature: Local test
    This feature file captures local test automation.
    Due to certain technical challenges in OpenShiftCI test infra
    we are keeping test scenario in local feature file for verifying the bits locally.
    Once the CI challenges are fixed, we move these test under basic tag.

    @local
    Scenario: Execute KAM bootstrap command with default and --push-to-git=true flag
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --git-host-access-token $GITHUB_TOKEN --push-to-git=true" succeeds
        Then stderr should be empty

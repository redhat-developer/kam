Feature: Manual test
    This feature file captures only manual test steps.
    Due to certain technical challenges in OpenShiftCI test infra
    we are keeping few test scenario in manual feature file.
    Once the challenges are fixed, we automate these manual steps too.

    @manual
    Scenario: Execute KAM bootstrap command with default flags
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --git-host-access-token $GITHUB_TOKEN" succeeds
        Then stderr should be empty
        And directory "gitops" should exist
Feature: Basic test
    Checks whether KAM top-level commands behave correctly.

    Scenario: KAM version
        When executing "kam version" succeeds
        Then stderr should be empty
        And stdout should contain "kam version"

    Scenario: KAM bootstrap command without --push-to-git=true flag
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GITHUB_TOKEN --output bootstrapresources" succeeds
        Then stderr should be empty
        Then executing "cd bootstrapresources"
        And executing "git init ."
        Then stderr should be empty
        And executing "git add ."
        Then stderr should be empty
        And executing "git commit -m "Initialcommit."
        Then stderr should be empty
        And executing "git remote add origin $GITOPS_REPO_URL"
        Then stderr should be empty
        And executing "git push -u origin master"
        Then stderr should be empty
        And executing "oc apply -k config/argocd/"
        Then stderr should be empty

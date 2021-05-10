@basic
Feature: Basic test
    Checks whether KAM top-level commands behave correctly.

    Scenario: KAM version
        When executing "kam version" succeeds
        Then stderr should be empty
        And stdout should match "kam\sversion\sv\d+\.\d+\.\d+"

    Scenario: Execute KAM bootstrap command without --push-to-git=true flag
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GIT_ACCESS_TOKEN --output bootstrapresources" succeeds
        Then stderr should be empty

    Scenario: Execute KAM bootstrap command that overwite the custom output manifest path
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GIT_ACCESS_TOKEN --output bootstrapresources --overwrite" succeeds
        Then stderr should be empty

    Scenario: KAM bootstrap command should fail if any one mandatory flag --git-host-access-token is missing
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL" fails
        Then exitcode should not equal "0"

    Scenario: Bringing the bootstrapped environment up
        Given gitops repository is created
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GIT_ACCESS_TOKEN --output bootstrapresources --overwrite" succeeds
        Then executing "cd bootstrapresources" succeeds
        Then executing "git init ." succeeds
        Then executing "git add ." succeeds
        Then executing "git commit -m 'Initial commit.'" succeeds
        Then executing "git remote add origin $GITOPS_REPO_URL" succeeds
        Then executing "git branch -M main" succeeds
        Then executing "git push -u origin main" succeeds
        Then executing "cd .." succeeds

    Scenario: Bringing the deployment infrastructure up
        Given gitops repository is created
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GIT_ACCESS_TOKEN --output bootstrapresources --overwrite" succeeds
        Then executing "cd bootstrapresources" succeeds
        Then executing "git init ." succeeds
        Then executing "git add ." succeeds
        Then executing "git commit -m 'Initial commit.'" succeeds
        Then executing "git remote add origin $GITOPS_REPO_URL" succeeds
        Then executing "git branch -M main" succeeds
        Then executing "git push -u origin main" succeeds
        Then executing "oc apply -k config/argocd/" succeeds
        Then login argocd API server
        


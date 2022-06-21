@gitlab
Feature: GitLab test
    Checks whether KAM top-level commands behave correctly.

    Scenario: Execute KAM bootstrap command without --overwrite flag
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GIT_ACCESS_TOKEN --output bootstrapresources" succeeds
        Then executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GIT_ACCESS_TOKEN --output bootstrapresources" fails
        And exitcode should not equal "0"

    Scenario: Execute KAM bootstrap command that overwrite the custom output manifest path
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GIT_ACCESS_TOKEN --output bootstrapresources" succeeds
        Then executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GIT_ACCESS_TOKEN --output bootstrapresources --overwrite" succeeds
        And stderr should be empty

    Scenario: KAM bootstrap command should fail if any one mandatory flag --git-host-access-token is missing
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL" fails
        Then exitcode should not equal "0"

    Scenario: Public GitLab repository without --push-to-git flag 
        Given "gitops" repository is created
        Given "taxi" repository exists
        Given "bus" repository exists
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GIT_ACCESS_TOKEN --output bootstrapresources" succeeds
        Then executing "cd bootstrapresources" succeeds
        And executing "git init ." succeeds
        And executing "git add ." succeeds
        And executing "git commit -m 'Initial commit.'" succeeds
        And executing "git remote add origin $GITOPS_REPO_URL'.git'" succeeds
        And executing "git branch -M main" succeeds
        And executing "git push -u origin main" succeeds
        Then executing "oc apply -k config/argocd/" succeeds
        Then wait for application "argo-app" to be in "Synced" state
        And wait for application "dev-app-taxi" to be in "Synced" state
        And wait for application "dev-env" to be in "Synced" state
        And wait for application "stage-env" to be in "Synced" state
        And wait for application "cicd-app" to be in "Synced" state
        Then executing "git checkout -b addNewService" succeeds
        And executing "oc apply -f ../secrets" succeeds
        When executing "kam webhook create --git-host-access-token $GIT_ACCESS_TOKEN --env-name dev --service-name taxi" succeeds
        Then stderr should be empty
        And executing "kam webhook create --git-host-access-token $GIT_ACCESS_TOKEN --cicd" succeeds
        Then stderr should be empty
        When executing "kam environment add --env-name new-env --pipelines-folder ." succeeds
        Then stderr should be empty
        When executing "kam service add --env-name new-env --app-name app-bus --service-name bus --git-repo-url $BUS_REPO_URL --pipelines-folder ." succeeds
        And executing "oc apply -f ../secrets/webhook-secret-new-env-bus.yaml" succeeds
        Then add kubernetes resource to the service in new environment
        Then executing "git add ." succeeds
        And executing "git commit -m 'Add new service'" succeeds
        And executing "git push origin addNewService" succeeds
        Then create and merge pull request after all the checks pass
        Then executing "git checkout main" succeeds
        And executing "git pull --rebase origin main" succeeds
        And executing "git push" succeeds
        Then wait for application "new-env-app-bus" to be in "Synced" state
        And wait for application "new-env-env" to be in "Synced" state
        And wait for application "argo-app" to be in "Synced" state
        And wait for application "dev-app-taxi" to be in "Synced" state
        And wait for application "dev-env" to be in "Synced" state
        And wait for application "stage-env" to be in "Synced" state
        And wait for application "cicd-app" to be in "Synced" state
        When executing "kam webhook create --git-host-access-token $GIT_ACCESS_TOKEN --env-name new-env --service-name bus --pipelines-folder ." succeeds
        Then stderr should be empty
        And executing "kam webhook delete --git-host-access-token $GIT_ACCESS_TOKEN --env-name new-env --service-name bus --pipelines-folder ." succeeds
        And executing "kam webhook delete --git-host-access-token $GIT_ACCESS_TOKEN --env-name dev --service-name taxi" succeeds
        And executing "oc delete -f ../secrets" succeeds
        And executing "oc delete ns cicd dev new-env stage" succeeds
        And executing "oc delete -k config/argocd/" succeeds
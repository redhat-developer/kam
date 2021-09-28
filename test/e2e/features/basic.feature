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

    Scenario: Bringing the bootstrapped environment up
        Given "gitops" repository is created
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GIT_ACCESS_TOKEN --output bootstrapresources" succeeds
        Then executing "cd bootstrapresources" succeeds
        And executing "git init ." succeeds
        And executing "git add ." succeeds
        And executing "git commit -m 'Initial commit.'" succeeds
        And executing "git remote add origin $GITOPS_REPO_URL" succeeds
        And executing "git branch -M main" succeeds
        And executing "git push -u origin main" succeeds
        And executing "cd .." succeeds

    Scenario: Bringing the deployment infrastructure up
        Given "gitops" repository is created
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GIT_ACCESS_TOKEN --output bootstrapresources" succeeds
        Then executing "cd bootstrapresources" succeeds
        And executing "git init ." succeeds
        And executing "git add ." succeeds
        And executing "git commit -m 'Initial commit.'" succeeds
        And executing "git remote add origin $GITOPS_REPO_URL" succeeds
        And executing "git branch -M main" succeeds
        And executing "git push -u origin main" succeeds
        Then executing "oc apply -k config/argocd/" succeeds
        And login argocd API server
        Then Wait for application "argo-app" to be in "Synced" state
        And Wait for application "dev-app-taxi" to be in "Synced" state
        And Wait for application "dev-env" to be in "Synced" state
        And Wait for application "stage-env" to be in "Synced" state
        And Wait for application "cicd-app" to be in "Synced" state
        Then executing "oc delete -k config/argocd" succeeds
        And executing "cd .." succeeds
        
    Scenario: First CI run
        Given "gitops" repository is created
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GIT_ACCESS_TOKEN --output bootstrapresources" succeeds
        Then executing "cd bootstrapresources" succeeds
        And executing "git init ." succeeds
        And executing "git add ." succeeds
        And executing "git commit -m 'Initial commit.'" succeeds
        And executing "git remote add origin $GITOPS_REPO_URL" succeeds
        And executing "git branch -M main" succeeds
        And executing "git push -u origin main" succeeds
        Then executing "oc apply -k config/argocd/" succeeds
        And login argocd API server
        Then Wait for application "argo-app" to be in "Synced" state
        And Wait for application "dev-app-taxi" to be in "Synced" state
        And Wait for application "dev-env" to be in "Synced" state
        And Wait for application "stage-env" to be in "Synced" state
        And Wait for application "cicd-app" to be in "Synced" state
        Then executing "cd .." succeeds
        And executing "oc apply -f secrets" succeeds
        And executing "cd bootstrapresources" succeeds
        When executing "kam webhook create --git-host-access-token $GIT_ACCESS_TOKEN --env-name dev --service-name taxi" succeeds
        Then stderr should be empty
        Then executing "kam webhook delete --git-host-access-token $GIT_ACCESS_TOKEN --env-name dev --service-name taxi" succeeds
        And executing "oc delete -k config/argocd" succeeds
        And executing "cd .." succeeds
        And executing "oc delete -f secrets" succeeds

    Scenario: Create an Application/Service in the new Environment and Commit and Push configuration to GitOps repository
        Given "gitops" repository is created
        And "bus" repository is created
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GIT_ACCESS_TOKEN --output bootstrapresources" succeeds
        Then executing "cd bootstrapresources" succeeds
        And executing "git init ." succeeds
        And executing "git add ." succeeds
        And executing "git commit -m 'Initial commit.'" succeeds
        And executing "git remote add origin $GITOPS_REPO_URL" succeeds
        And executing "git branch -M main" succeeds
        And executing "git push -u origin main" succeeds
        Then executing "oc apply -k config/argocd/" succeeds
        And login argocd API server
        Then Wait for application "argo-app" to be in "Synced" state
        And Wait for application "dev-app-taxi" to be in "Synced" state
        And Wait for application "dev-env" to be in "Synced" state
        And Wait for application "stage-env" to be in "Synced" state
        And Wait for application "cicd-app" to be in "Synced" state
        Then executing "cd .." succeeds
        And executing "oc apply -f secrets" succeeds
        And executing "cd bootstrapresources" succeeds
        When executing "kam webhook create --git-host-access-token $GIT_ACCESS_TOKEN --env-name dev --service-name taxi" succeeds
        Then stderr should be empty
        And executing "cd .." succeeds
        When executing "kam environment add --env-name new-env --pipelines-folder bootstrapresources" succeeds
        Then stderr should be empty
        When executing "kam service add --env-name new-env --app-name app-bus --service-name bus --git-repo-url $BUS_REPO_URL --pipelines-folder bootstrapresources" succeeds
        And executing "oc apply -f secrets/webhook-secret-new-env-bus.yaml" succeeds
        Then add kubernetes resource to newly created config file
        And executing "cd bootstrapresources" succeeds
        Then executing "git add ." succeeds
        And executing "git commit -m 'Add new service'" succeeds
        And executing "git push origin main" succeeds
        Then Wait for application "new-env-app-bus" to deploy
        And Wait for application "new-env-env" to deploy
        Then Wait for application "argo-app" to be in "Synced" state
        And Wait for application "dev-app-taxi" to be in "Synced" state
        And Wait for application "dev-env" to be in "Synced" state
        And Wait for application "stage-env" to be in "Synced" state
        And Wait for application "cicd-app" to be in "Synced" state
        And Wait for application "new-env-app-bus" to be in "Synced" state
        Then executing "kam webhook delete --git-host-access-token $GIT_ACCESS_TOKEN --env-name dev --service-name taxi" succeeds
        And executing "oc delete -k config/argocd" succeeds
        And executing "cd .." succeeds
        And executing "oc delete -f secrets" succeeds

    Scenario: Create Webhook
        Given "gitops" repository is created
        And "bus" repository is created
        When executing "kam bootstrap --service-repo-url $SERVICE_REPO_URL --gitops-repo-url $GITOPS_REPO_URL --image-repo $IMAGE_REPO --dockercfgjson $DOCKERCONFIGJSON_PATH --git-host-access-token $GIT_ACCESS_TOKEN --output bootstrapresources" succeeds
        Then executing "cd bootstrapresources" succeeds
        And executing "git init ." succeeds
        And executing "git add ." succeeds
        And executing "git commit -m 'Initial commit.'" succeeds
        And executing "git remote add origin $GITOPS_REPO_URL" succeeds
        And executing "git branch -M main" succeeds
        And executing "git push -u origin main" succeeds
        Then executing "oc apply -k config/argocd/" succeeds
        And login argocd API server
        Then Wait for application "argo-app" to be in "Synced" state
        And Wait for application "dev-app-taxi" to be in "Synced" state
        And Wait for application "dev-env" to be in "Synced" state
        And Wait for application "stage-env" to be in "Synced" state
        And Wait for application "cicd-app" to be in "Synced" state
        Then executing "cd .." succeeds
        And executing "oc apply -f secrets" succeeds
        And executing "cd bootstrapresources" succeeds
        When executing "kam webhook create --git-host-access-token $GIT_ACCESS_TOKEN --env-name dev --service-name taxi" succeeds
        Then stderr should be empty
        And executing "cd .." succeeds
        When executing "kam environment add --env-name new-env --pipelines-folder bootstrapresources" succeeds
        Then stderr should be empty
        When executing "kam service add --env-name new-env --app-name app-bus --service-name bus --git-repo-url $BUS_REPO_URL --pipelines-folder bootstrapresources" succeeds
        And executing "oc apply -f secrets/webhook-secret-new-env-bus.yaml" succeeds
        Then add kubernetes resource to newly created config file
        And executing "cd bootstrapresources" succeeds
        Then executing "git add ." succeeds
        And executing "git commit -m 'Add new service'" succeeds
        And executing "git push origin main" succeeds
        Then Wait for application "new-env-app-bus" to deploy
        And Wait for application "new-env-env" to deploy
        Then Wait for application "argo-app" to be in "Synced" state
        And Wait for application "dev-app-taxi" to be in "Synced" state
        And Wait for application "dev-env" to be in "Synced" state
        And Wait for application "stage-env" to be in "Synced" state
        And Wait for application "cicd-app" to be in "Synced" state
        And Wait for application "new-env-app-bus" to be in "Synced" state
        Then executing "cd .." succeeds
        When executing "kam webhook create --git-host-access-token $GIT_ACCESS_TOKEN --env-name new-env --service-name bus --pipelines-folder bootstrapresources" succeeds
        Then stderr should be empty
        And executing "kam webhook delete --git-host-access-token $GIT_ACCESS_TOKEN --env-name new-env --service-name bus --pipelines-folder bootstrapresources" succeeds
        And executing "oc delete -f secrets" succeeds
        And executing "cd bootstrapresources" succeeds
        And executing "kam webhook delete --git-host-access-token $GIT_ACCESS_TOKEN --env-name dev --service-name taxi" succeeds
        And executing "oc delete -k config/argocd" succeeds
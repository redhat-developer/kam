## Steps to follow for seting up a sample GitOps pipelines using [kam](../../releases)

- Complete all the pre-requisite steps mentioned in the [docs](/docs/journey/day1#day-1-operations)
- Bootstrap a GitOps repository following steps mentioned in the [docs](/docs/journey/day1#bootstrapping-the-manifest)

  ```sh
  $ kam bootstrap --service-repo-url https://github.com/<username>/taxi.git \          
    --dockercfgjson ~/Downloads/<username>-gitops-auth.json \
    --gitops-repo-url https://github.com/<username>/gitops-example.git \
    --image-repo quay.io/<username>/taxi --prefix demo --output <output directory> \
    --sealed-secrets-svc sealed-secrets-controller --sealed-secrets-ns kube-system

  ```

- This bootstraps your GitOps repo with two environments `demo-dev` and `demo-stage`.
  - `demo-dev` has an app `app-taxi` with one service `taxi`.
- Update the labels of the deployment of taxi service.
  - Copy labels from `environments/demo-dev/apps/app-taxi/services/taxi/base/config/200-service.yaml` to `environments/demo-dev/apps/app-taxi/services/taxi/base/config/100-deployment.yaml` 
- Add a new service `bus` to the created app `app-taxi` in `demo-dev` environment - 

  ```sh
  $ kam service add  --app-name app-taxi --env-name demo-dev \
    --pipelines-folder <bootstrapped gitops folder> \
    --service-name bus --sealed-secrets-ns kube-system \
    --sealed-secrets-svc sealed-secrets-controller \
    --git-repo-url https://github.com/<username>/bus.git
  ```

- Copy `config` folder from `taxi/base/` to `bus/base/`.
  - Update `100-deployment.yaml` and `200-service.yaml` to change names and labels from `taxi` to `bus`.

- Add `taxi` and `bus` service to `app-taxi` for `demo-stage` environment.

  ```sh
  $ kam service add  --app-name app-taxi --env-name demo-stage \
    --pipelines-folder <bootstrapped gitops folder> \
    --service-name taxi --sealed-secrets-ns kube-system \
    --sealed-secrets-svc sealed-secrets-controller

  ```

  ```sh
  $ kam service add  --app-name app-taxi --env-name demo-stage \
    --pipelines-folder <bootstrapped gitops folder> \
    --service-name bus --sealed-secrets-ns kube-system \
    --sealed-secrets-svc sealed-secrets-controller

  ```

- Copy `config` folder from both the services `taxi` and `bus` of `demo-dev` to `demo-stage`.
  - Update the namespace of service and deployments of `demo-stage`.
  - Update the replicas for deployments to 2.

- Add a new environement `demo-prod`.
  
  ```sh
  $ kam environment add --env-name demo-prod --pipelines-folder  <bootstrapped gitops folder>
  ```

- Add `taxi` and `bus` service to `app-taxi` for `demo-prod` environment.

  ```sh
  $ kam service add  --app-name app-taxi --env-name demo-prod \
    --pipelines-folder <bootstrapped gitops folder> \
    --service-name taxi --sealed-secrets-ns kube-system \
    --sealed-secrets-svc sealed-secrets-controller

  ```

  ```sh
  $ kam service add  --app-name app-taxi --env-name demo-prod \
    --pipelines-folder <bootstrapped gitops folder> \
    --service-name bus --sealed-secrets-ns kube-system \
    --sealed-secrets-svc sealed-secrets-controller

  ```

- Copy `config` folder from both the services `taxi` and `bus` of `demo-dev` to `demo-prod`.
  - Update the namespace of service and deployments of `demo-prod`.
  - Update the replicas for deployments to 3.
- Bring up the deployments and environments - 
  - `oc apply -k config/argocd/`
  - `oc apply -k config/demo-cicd/base`
  - `oc apply -k environments/demo-dev/env/base`
  - `oc apply -k environments/demo-stage/env/base`
  - `oc apply -k environments/demo-prod/env/base`
  - `oc apply -k environments/demo-dev/apps/app-taxi`
  - `oc apply -k environments/demo-stage/apps/app-taxi`
  - `oc apply -k environments/demo-prod/apps/app-taxi`
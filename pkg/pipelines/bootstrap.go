package pipelines

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	"github.com/mitchellh/go-homedir"
	"github.com/openshift/odo/pkg/log"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
	v1rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/redhat-developer/kam/pkg/pipelines/config"
	"github.com/redhat-developer/kam/pkg/pipelines/deployment"
	"github.com/redhat-developer/kam/pkg/pipelines/dryrun"
	"github.com/redhat-developer/kam/pkg/pipelines/eventlisteners"
	"github.com/redhat-developer/kam/pkg/pipelines/imagerepo"
	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
	"github.com/redhat-developer/kam/pkg/pipelines/meta"
	"github.com/redhat-developer/kam/pkg/pipelines/namespaces"
	"github.com/redhat-developer/kam/pkg/pipelines/pipelines"
	res "github.com/redhat-developer/kam/pkg/pipelines/resources"
	"github.com/redhat-developer/kam/pkg/pipelines/roles"
	"github.com/redhat-developer/kam/pkg/pipelines/routes"
	"github.com/redhat-developer/kam/pkg/pipelines/scm"
	"github.com/redhat-developer/kam/pkg/pipelines/secrets"
	"github.com/redhat-developer/kam/pkg/pipelines/statustracker"
	"github.com/redhat-developer/kam/pkg/pipelines/tasks"
	"github.com/redhat-developer/kam/pkg/pipelines/triggers"
	"github.com/redhat-developer/kam/pkg/pipelines/yaml"
)

const (
	// Kustomize constants for kustomization.yaml
	Kustomize = "kustomization.yaml"

	namespacesPath        = "01-namespaces/cicd-environment.yaml"
	rolesPath             = "02-rolebindings/pipeline-service-role.yaml"
	rolebindingsPath      = "02-rolebindings/pipeline-service-rolebinding.yaml"
	serviceAccountPath    = "02-rolebindings/pipeline-service-account.yaml"
	secretsPath           = "03-secrets/gitops-webhook-secret.yaml"     //nolint:gosec
	authTokenPath         = "03-secrets/git-host-access-token.yaml"     // nolint:gosec
	basicAuthTokenPath    = "03-secrets/git-host-basic-auth-token.yaml" // nolint:gosec
	dockerConfigPath      = "03-secrets/docker-config.yaml"
	gitopsTasksPath       = "04-tasks/deploy-from-source-task.yaml"
	ciPipelinesPath       = "05-pipelines/ci-dryrun-from-push-pipeline.yaml"
	appCiPipelinesPath    = "05-pipelines/app-ci-pipeline.yaml"
	pushTemplatePath      = "07-templates/ci-dryrun-from-push-template.yaml"
	appCIPushTemplatePath = "07-templates/app-ci-build-from-push-template.yaml"
	eventListenerPath     = "08-eventlisteners/cicd-event-listener.yaml"
	routePath             = "09-routes/gitops-webhook-event-listener.yaml"

	dockerSecretName = "regcred"

	saName              = "pipeline"
	roleBindingName     = "pipelines-service-role-binding"
	webhookSecretLength = 20

	pipelinesFile     = "pipelines.yaml"
	bootstrapImage    = "nginxinc/nginx-unprivileged:latest"
	appCITemplateName = "app-ci-template"
	version           = 1
)

// BootstrapOptions is a struct that provides the optional flags
type BootstrapOptions struct {
	GitOpsRepoURL            string // This is where the pipelines and configuration are.
	GitOpsWebhookSecret      string // This is the secret for authenticating hooks from your GitOps repo.
	Prefix                   string
	DockerConfigJSONFilename string
	ImageRepo                string               // This is where built images are pushed to.
	OutputPath               string               // Where to write the bootstrapped files to?
	SealedSecretsService     types.NamespacedName // SealedSecrets Services name
	GitHostAccessToken       string               // The auth token to use to send commit-status notifications, and access private repositories.
	Overwrite                bool                 // This allows to overwrite if there is an existing gitops repository
	ServiceRepoURL           string               // This is the full URL to your GitHub repository for your app source.
	SaveTokenKeyRing         bool                 // If true, the access-token will be saved in the keyring
	ServiceWebhookSecret     string               // This is the secret for authenticating hooks from your app source.
	PrivateRepoDriver        string               // Records the type of the GitOpsRepoURL driver if not a well-known host.
	CommitStatusTracker      bool                 // If true, this is a "private repository", i.e. requires authentication to clone the repository.
}

// PolicyRules to be bound to service account
var (
	Rules = []v1rbac.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"namespaces", "services"},
			Verbs:     []string{"patch", "get", "create"},
		},
		{
			APIGroups: []string{"rbac.authorization.k8s.io"},
			Resources: []string{"clusterroles", "roles"},
			Verbs:     []string{"bind", "patch", "get"},
		},
		{
			APIGroups: []string{"rbac.authorization.k8s.io"},
			Resources: []string{"clusterrolebindings", "rolebindings"},
			Verbs:     []string{"get", "create", "patch"},
		},
		{
			APIGroups: []string{"bitnami.com"},
			Resources: []string{"sealedsecrets"},
			Verbs:     []string{"get", "patch", "create"},
		},
		{
			APIGroups: []string{"apps"},
			Resources: []string{"deployments"},
			Verbs:     []string{"get", "create", "patch"},
		},
		{
			APIGroups: []string{"argoproj.io"},
			Resources: []string{"applications", "argocds"},
			Verbs:     []string{"get", "create", "patch"},
		},
	}
)

// Bootstrap is the entry-point from the CLI for bootstrapping the GitOps
// configuration.
func Bootstrap(o *BootstrapOptions, appFs afero.Fs) error {
	err := checkPipelinesFileExists(appFs, o.OutputPath, o.Overwrite)
	if err != nil {
		return err
	}
	err = maybeMakeHookSecrets(o)
	if err != nil {
		return err
	}

	bootstrapped, err := bootstrapResources(o, appFs)
	if err != nil {
		return fmt.Errorf("failed to bootstrap resources: %v", err)
	}

	m := bootstrapped[pipelinesFile].(*config.Manifest)
	built, err := buildResources(appFs, m)
	if err != nil {
		return fmt.Errorf("failed to build resources: %v", err)
	}

	bootstrapped = res.Merge(built, bootstrapped)
	log.Successf("Created dev, stage and CICD environments")
	_, err = yaml.WriteResources(appFs, o.OutputPath, bootstrapped)
	if err != nil {
		return fmt.Errorf("failed to write resources: %w", err)
	}

	return nil
}

func maybeMakeHookSecrets(o *BootstrapOptions) error {
	if o.GitOpsWebhookSecret == "" {
		gitopsSecret, err := secrets.GenerateString(webhookSecretLength)
		if err != nil {
			return fmt.Errorf("failed to generate GitOps webhook secret: %v", err)
		}
		o.GitOpsWebhookSecret = gitopsSecret
	}
	if o.ServiceWebhookSecret == "" {
		appSecret, err := secrets.GenerateString(webhookSecretLength)
		if err != nil {
			return fmt.Errorf("failed to generate application webhook secret: %v", err)
		}
		o.ServiceWebhookSecret = appSecret
	}
	return nil
}

func bootstrapResources(o *BootstrapOptions, appFs afero.Fs) (res.Resources, error) {
	isInternalRegistry, imageRepo, err := imagerepo.ValidateImageRepo(o.ImageRepo)
	if err != nil {
		return nil, err
	}
	gitOpsRepo, err := scm.NewRepository(o.GitOpsRepoURL)
	if err != nil {
		return nil, err
	}
	appRepo, err := scm.NewRepository(o.ServiceRepoURL)
	if err != nil {
		return nil, err
	}
	repoName, err := repoFromURL(appRepo.URL())
	if err != nil {
		return nil, fmt.Errorf("invalid app repo URL: %v", err)
	}

	bootstrapped, err := createInitialFiles(
		appFs, gitOpsRepo, o)
	if err != nil {
		return nil, err
	}
	appName := repoToAppName(repoName)
	serviceName := repoName
	ns := namespaces.NamesWithPrefix(o.Prefix)
	secretName := secrets.MakeServiceWebhookSecretName(ns["dev"], serviceName)
	envs, configEnv, err := bootstrapEnvironments(appRepo, o.Prefix, secretName, ns)
	if err != nil {
		return nil, err
	}
	if o.PrivateRepoDriver != "" {
		host, err := scm.HostnameFromURL(o.GitOpsRepoURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get hostname from URL %q: %w", o.GitOpsRepoURL, err)
		}
		configEnv.Git = &config.GitConfig{Drivers: map[string]string{host: o.PrivateRepoDriver}}
	}
	m := createManifest(gitOpsRepo.URL(), configEnv, envs...)

	devEnv := m.GetEnvironment(ns["dev"])
	if devEnv == nil {
		return nil, errors.New("unable to bootstrap without dev environment")
	}

	app := m.GetApplication(ns["dev"], appName)
	if app == nil {
		return nil, errors.New("unable to bootstrap without application")
	}
	svcFiles, err := bootstrapServiceDeployment(devEnv, app)
	if err != nil {
		return nil, fmt.Errorf("failed to create bootstrap service: %w", err)
	}
	hookSecret, err := secrets.CreateSealedSecret(
		meta.NamespacedName(ns["cicd"], secretName),
		o.SealedSecretsService,
		o.ServiceWebhookSecret,
		eventlisteners.WebhookSecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GitHub Webhook Secret: %v", err)
	}

	cfg := m.GetPipelinesConfig()
	if cfg == nil {
		return nil, errors.New("failed to find a pipeline configuration - unable to continue bootstrap")
	}
	secretFilename := filepath.Join("03-secrets", secretName+".yaml")
	secretsPath := filepath.Join(config.PathForPipelines(cfg), "base", secretFilename)
	bootstrapped[secretsPath] = hookSecret

	bindingName, imageRepoBindingFilename, svcImageBinding := createSvcImageBinding(cfg, devEnv, appName, serviceName, imageRepo, !isInternalRegistry)
	bootstrapped = res.Merge(svcImageBinding, bootstrapped)

	kustomizePath := filepath.Join(config.PathForPipelines(cfg), "base", "kustomization.yaml")
	k, ok := bootstrapped[kustomizePath].(res.Kustomization)
	if !ok {
		return nil, fmt.Errorf("no kustomization for the %s environment found", kustomizePath)
	}
	if isInternalRegistry {
		filenames, resources, err := imagerepo.CreateInternalRegistryResources(
			cfg, roles.CreateServiceAccount(meta.NamespacedName(cfg.Name, saName)),
			imageRepo, o.GitOpsRepoURL)
		if err != nil {
			return nil, fmt.Errorf("failed to get resources for internal image repository: %v", err)
		}
		bootstrapped = res.Merge(resources, bootstrapped)
		k.Resources = append(k.Resources, filenames...)
	}

	// This is specific to bootstrap, because there's only one service.
	devEnv.Apps[0].Services[0].Pipelines = &config.Pipelines{
		Integration: &config.TemplateBinding{
			Bindings: append([]string{bindingName}, devEnv.Pipelines.Integration.Bindings...),
		},
	}
	bootstrapped[pipelinesFile] = m

	k.Resources = append(k.Resources, secretFilename, imageRepoBindingFilename)
	sort.Strings(k.Resources)
	bootstrapped[kustomizePath] = k

	bootstrapped = res.Merge(svcFiles, bootstrapped)
	return bootstrapped, nil
}

func bootstrapServiceDeployment(dev *config.Environment, app *config.Application) (res.Resources, error) {
	svc := dev.Apps[0].Services[0]
	svcBase := filepath.Join(config.PathForService(app, dev, svc.Name), "base", "config")
	resources := res.Resources{}
	// TODO: This should change if we add Namespace to Environment.
	// We'd need to create the resources in the namespace _of_ the Environment.
	resources[filepath.Join(svcBase, "100-deployment.yaml")] = deployment.Create(app.Name, dev.Name, svc.Name, bootstrapImage, deployment.ContainerPort(8080))
	containerSvc := createBootstrapService(app.Name, dev.Name, svc.Name)
	resources[filepath.Join(svcBase, "200-service.yaml")] = containerSvc
	r, err := routes.NewFromService(containerSvc)
	if err != nil {
		return nil, err
	}
	resources[filepath.Join(svcBase, "300-route.yaml")] = r
	resources[filepath.Join(svcBase, "kustomization.yaml")] = &res.Kustomization{
		Resources: []string{
			"100-deployment.yaml",
			"200-service.yaml",
			"300-route.yaml",
		}}
	return resources, nil
}

func bootstrapEnvironments(repo scm.Repository, prefix, secretName string, ns map[string]string) ([]*config.Environment, *config.Config, error) {
	envs := []*config.Environment{}
	var pipelinesConfig *config.PipelinesConfig
	for _, k := range []string{"cicd", "dev", "stage"} {
		v := ns[k]
		if k == "cicd" {
			pipelinesConfig = &config.PipelinesConfig{Name: prefix + "cicd"}
		} else {
			env := &config.Environment{Name: v}
			if k == "dev" {
				svc, err := serviceFromRepo(repo.URL(), secretName, ns["cicd"])
				if err != nil {
					return nil, nil, err
				}
				app, err := applicationFromRepo(repo.URL(), svc)
				if err != nil {
					return nil, nil, err
				}
				app.Services = []*config.Service{svc}
				env.Apps = []*config.Application{app}
				env.Pipelines = defaultPipelines(repo)
			}
			envs = append(envs, env)
		}
	}
	cfg := &config.Config{Pipelines: pipelinesConfig, ArgoCD: &config.ArgoCDConfig{Namespace: "argocd"}}
	return envs, cfg, nil
}

func serviceFromRepo(repoURL, secretName, secretNS string) (*config.Service, error) {
	repo, err := repoFromURL(repoURL)
	if err != nil {
		return nil, err
	}
	return &config.Service{
		Name:      repo,
		SourceURL: repoURL,
		Webhook: &config.Webhook{
			Secret: &config.Secret{
				Name:      secretName,
				Namespace: secretNS,
			},
		},
	}, nil
}

func applicationFromRepo(repoURL string, service *config.Service) (*config.Application, error) {
	repo, err := repoFromURL(repoURL)
	if err != nil {
		return nil, err
	}
	return &config.Application{
		Name:     repoToAppName(repo),
		Services: []*config.Service{service},
	}, nil
}

func repoFromURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	parts := strings.Split(u.Path, "/")
	return strings.TrimSuffix(parts[len(parts)-1], ".git"), nil
}

func orgRepoFromURL(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	parts := strings.Split(u.Path, "/")
	orgRepo := strings.Join(parts[len(parts)-2:], "/")
	return strings.TrimSuffix(orgRepo, ".git"), nil
}

func createBootstrapService(appName, ns, name string) *corev1.Service {
	svc := &corev1.Service{
		TypeMeta:   meta.TypeMeta("Service", "v1"),
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, name)),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       8080,
					TargetPort: intstr.FromInt(8080)},
			},
		},
	}
	labels := map[string]string{
		deployment.KubernetesAppNameLabel: name,
		deployment.KubernetesPartOfLabel:  appName,
	}
	svc.ObjectMeta.Labels = labels
	svc.Spec.Selector = labels
	return svc
}

func repoToAppName(repoName string) string {
	return "app-" + repoName
}

func defaultPipelines(r scm.Repository) *config.Pipelines {
	return &config.Pipelines{
		Integration: &config.TemplateBinding{
			Template: appCITemplateName,
			Bindings: []string{r.PushBindingName()},
		},
	}
}

// Checks whether the pipelines.yaml is present in the output path specified.
func checkPipelinesFileExists(appFs afero.Fs, outputPath string, overWrite bool) error {
	exists, _ := ioutils.IsExisting(appFs, filepath.Join(outputPath, pipelinesFile))
	if exists && !overWrite {
		return fmt.Errorf("pipelines.yaml in output path already exists. If you want to replace your existing files, please rerun with --overwrite")
	}
	return nil
}

func createInitialFiles(fs afero.Fs, repo scm.Repository, o *BootstrapOptions) (res.Resources, error) {
	cicd := &config.PipelinesConfig{Name: o.Prefix + "cicd"}
	pipelineConfig := &config.Config{Pipelines: cicd}
	manifest := createManifest(repo.URL(), pipelineConfig)
	initialFiles := res.Resources{
		pipelinesFile: manifest,
	}
	resources, err := createCICDResources(fs, repo, cicd, o)
	if err != nil {
		return nil, err
	}

	files := getResourceFiles(resources)
	prefixedResources := addPrefixToResources(pipelinesPath(manifest.Config), resources)
	initialFiles = res.Merge(prefixedResources, initialFiles)

	pipelinesConfigKustomizations := addPrefixToResources(
		config.PathForPipelines(manifest.Config.Pipelines),
		getCICDKustomization(files))
	initialFiles = res.Merge(pipelinesConfigKustomizations, initialFiles)

	return initialFiles, nil
}

// createDockerSecret creates a secret that allows pushing images to upstream repositories.
func createDockerSecret(fs afero.Fs, dockerConfigJSONFilename, secretNS string, sealedSecretsService types.NamespacedName) (*ssv1alpha1.SealedSecret, error) {
	if dockerConfigJSONFilename == "" {
		return nil, errors.New("failed to generate path to file: --dockerconfigjson flag is not provided")
	}
	authJSONPath, err := homedir.Expand(dockerConfigJSONFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to generate path to file: %v", err)
	}
	f, err := fs.Open(authJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Docker config %#v : %s", authJSONPath, err)
	}
	defer f.Close()

	dockerSecret, err := secrets.CreateSealedDockerConfigSecret(meta.NamespacedName(secretNS, dockerSecretName), sealedSecretsService, f)
	if err != nil {
		return nil, err
	}

	return dockerSecret, nil
}

// createCICDResources creates resources for OpenShift pipelines.
func createCICDResources(fs afero.Fs, repo scm.Repository, pipelineConfig *config.PipelinesConfig, o *BootstrapOptions) (res.Resources, error) {
	cicdNamespace := pipelineConfig.Name
	// key: path of the resource
	// value: YAML content of the resource
	outputs := map[string]interface{}{}
	githubSecret, err := secrets.CreateSealedSecret(meta.NamespacedName(cicdNamespace, eventlisteners.GitOpsWebhookSecret),
		o.SealedSecretsService, o.GitOpsWebhookSecret, eventlisteners.WebhookSecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate GitHub Webhook Secret: %w", err)
	}
	outputs[secretsPath] = githubSecret
	outputs[namespacesPath] = namespaces.Create(cicdNamespace, o.GitOpsRepoURL)
	outputs[rolesPath] = roles.CreateClusterRole(meta.NamespacedName("", roles.ClusterRoleName), Rules)

	sa := roles.CreateServiceAccount(meta.NamespacedName(cicdNamespace, saName))

	if o.DockerConfigJSONFilename != "" {
		dockerSecret, err := createDockerSecret(fs, o.DockerConfigJSONFilename, cicdNamespace,
			o.SealedSecretsService)
		if err != nil {
			return nil, err
		}
		outputs[dockerConfigPath] = dockerSecret
		log.Success("Authentication tokens encrypted in secrets")
		outputs[serviceAccountPath] = roles.AddSecretToSA(sa, dockerSecretName)
	}

	if o.GitHostAccessToken != "" {
		err := generateSecrets(outputs, sa, cicdNamespace, o)
		if err != nil {
			return nil, err
		}
	}

	if o.CommitStatusTracker {
		trackerResources, err := statustracker.Resources(cicdNamespace, o.GitOpsRepoURL, o.PrivateRepoDriver)
		if err != nil {
			return nil, err
		}
		outputs = res.Merge(outputs, trackerResources)
		log.Success("Pipelines tracker has been configured")
	}

	outputs[rolebindingsPath] = roles.CreateClusterRoleBinding(meta.NamespacedName("", roleBindingName), sa, "ClusterRole", roles.ClusterRoleName)
	script, err := dryrun.MakeScript("kubectl", cicdNamespace)
	if err != nil {
		return nil, err
	}
	outputs[gitopsTasksPath] = tasks.CreateDeployFromSourceTask(cicdNamespace, script)
	outputs[ciPipelinesPath] = pipelines.CreateCIPipeline(meta.NamespacedName(cicdNamespace, "ci-dryrun-from-push-pipeline"), cicdNamespace)
	outputs[appCiPipelinesPath] = pipelines.CreateAppCIPipeline(meta.NamespacedName(cicdNamespace, "app-ci-pipeline"))
	pushBinding, pushBindingName := repo.CreatePushBinding(cicdNamespace)
	outputs[filepath.Join("06-bindings", pushBindingName+".yaml")] = pushBinding
	outputs[pushTemplatePath] = triggers.CreateCIDryRunTemplate(cicdNamespace, saName)
	outputs[appCIPushTemplatePath] = triggers.CreateDevCIBuildPRTemplate(cicdNamespace, saName)
	outputs[eventListenerPath] = eventlisteners.Generate(repo, cicdNamespace, saName, eventlisteners.GitOpsWebhookSecret)
	log.Success("OpenShift Pipelines resources created")
	route, err := eventlisteners.GenerateRoute(cicdNamespace)
	if err != nil {
		return nil, err
	}
	outputs[routePath] = route
	log.Success("Openshift Route for EventListener created")
	return outputs, nil
}

func createManifest(gitOpsRepoURL string, configEnv *config.Config, envs ...*config.Environment) *config.Manifest {
	return &config.Manifest{
		GitOpsURL:    gitOpsRepoURL,
		Environments: envs,
		Config:       configEnv,
		Version:      version,
	}
}

func getCICDKustomization(files []string) res.Resources {
	return res.Resources{
		"overlays/kustomization.yaml": res.Kustomization{
			Bases: []string{"../base"},
		},
		"base/kustomization.yaml": res.Kustomization{
			Resources: files,
		},
	}
}

func pipelinesPath(m *config.Config) string {
	return filepath.Join(config.PathForPipelines(m.Pipelines), "base")
}

func addPrefixToResources(prefix string, files res.Resources) map[string]interface{} {
	updated := map[string]interface{}{}
	for k, v := range files {
		updated[filepath.Join(prefix, k)] = v
	}
	return updated
}

func getResourceFiles(r res.Resources) []string {
	files := []string{}
	for k := range r {
		files = append(files, k)
	}
	sort.Strings(files)
	return files
}

func generateSecrets(outputs res.Resources, sa *corev1.ServiceAccount, ns string, o *BootstrapOptions) error {
	if o.CommitStatusTracker {
		tokenSecret, err := secrets.CreateSealedSecret(meta.NamespacedName(
			ns, statustracker.CommitStatusTrackerSecret), o.SealedSecretsService, o.GitHostAccessToken, "token")
		if err != nil {
			return fmt.Errorf("failed to generate access token Secret: %w", err)
		}
		outputs[authTokenPath] = tokenSecret
		outputs[serviceAccountPath] = roles.AddSecretToSA(sa, tokenSecret.Name)
	}
	secretTargetHost, err := repoURL(o.ServiceRepoURL)
	if err != nil {
		return fmt.Errorf("failed to parse the Service Repo URL %q: %w", o.ServiceRepoURL, err)
	}
	basicAuthSecret, err := secrets.CreateSealedBasicAuthSecret(meta.NamespacedName(
		ns, "git-host-basic-auth-token"), o.SealedSecretsService, o.GitHostAccessToken, meta.AddAnnotations(map[string]string{
		"tekton.dev/git-0": secretTargetHost,
	}))
	if err != nil {
		return fmt.Errorf("failed to generate basic auth token Secret: %w", err)
	}
	outputs[basicAuthTokenPath] = basicAuthSecret
	outputs[serviceAccountPath] = roles.AddSecretToSA(sa, basicAuthSecret.Name)
	return nil
}

package pipelines

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/config"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/environments"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/eventlisteners"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/imagerepo"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/meta"
	res "github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/resources"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/roles"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/secrets"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/triggers"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/yaml"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/types"
)

// AddServiceOptions control how new services are added to the configuration.
type AddServiceOptions struct {
	AppName                  string
	EnvName                  string
	GitRepoURL               string
	ImageRepo                string
	InternalRegistryHostname string
	PipelinesFolderPath      string
	ServiceName              string
	WebhookSecret            string
	SealedSecretsService     types.NamespacedName // SealedSecrets service name
}

func AddService(o *AddServiceOptions, appFs afero.Fs) error {
	m, err := config.LoadManifest(appFs, o.PipelinesFolderPath)
	if err != nil {
		return err
	}
	files, err := serviceResources(m, appFs, o)
	if err != nil {
		return err
	}

	_, err = yaml.WriteResources(appFs, o.PipelinesFolderPath, files)
	if err != nil {
		return err
	}
	cfg := m.GetPipelinesConfig()
	if cfg != nil {
		base := filepath.Join(o.PipelinesFolderPath, config.PathForPipelines(cfg), "base")
		err = updateKustomization(appFs, base)
		if err != nil {
			return err
		}
	}
	return nil
}

func serviceResources(m *config.Manifest, appFs afero.Fs, o *AddServiceOptions) (res.Resources, error) {
	files := res.Resources{}
	svc, err := createService(o.ServiceName, o.GitRepoURL)
	if err != nil {
		return nil, err
	}
	cfg := m.GetPipelinesConfig()
	if cfg != nil && o.WebhookSecret == "" && o.GitRepoURL != "" {
		gitSecret, err := secrets.GenerateString(webhookSecretLength)
		if err != nil {
			return nil, fmt.Errorf("failed to generate service webhook secret: %v", err)
		}
		o.WebhookSecret = gitSecret
	}

	env := m.GetEnvironment(o.EnvName)
	if env == nil {
		return nil, fmt.Errorf("environment %s does not exist", o.EnvName)
	}

	// add the secret only if CI/CD env is present
	if cfg != nil {
		secretName := secrets.MakeServiceWebhookSecretName(o.EnvName, svc.Name)
		hookSecret, err := secrets.CreateSealedSecret(
			meta.NamespacedName(cfg.Name, secretName), o.SealedSecretsService, o.WebhookSecret,
			eventlisteners.WebhookSecretKey)
		if err != nil {
			return nil, err
		}

		svc.Webhook = &config.Webhook{
			Secret: &config.Secret{
				Name:      secretName,
				Namespace: cfg.Name,
			},
		}
		secretFilename := filepath.Join("03-secrets", secretName+".yaml")
		secretsPath := filepath.Join(config.PathForPipelines(cfg), "base", secretFilename)
		files[secretsPath] = hookSecret

		if o.ImageRepo != "" {
			_, resources, bindingName, err := createImageRepoResources(m, cfg, env, o)
			if err != nil {
				return nil, err
			}

			files = res.Merge(resources, files)
			svc.Pipelines = &config.Pipelines{
				Integration: &config.TemplateBinding{
					Bindings: append([]string{bindingName}, env.Pipelines.Integration.Bindings[:]...),
				},
			}
		}
	}

	err = m.AddService(o.EnvName, o.AppName, svc)
	if err != nil {
		return nil, err
	}
	err = m.Validate()
	if err != nil {
		return nil, err
	}

	files[filepath.Base(filepath.Join(o.PipelinesFolderPath, pipelinesFile))] = m
	buildParams := &BuildParameters{
		PipelinesFolderPath: o.PipelinesFolderPath,
		OutputPath:          o.PipelinesFolderPath,
	}
	built, err := buildResources(appFs, buildParams, m)
	if err != nil {
		return nil, err
	}
	return res.Merge(built, files), nil
}

func createImageRepoResources(m *config.Manifest, cfg *config.PipelinesConfig, env *config.Environment, p *AddServiceOptions) ([]string, res.Resources, string, error) {
	isInternalRegistry, imageRepo, err := imagerepo.ValidateImageRepo(p.ImageRepo, p.InternalRegistryHostname)
	if err != nil {
		return nil, nil, "", err
	}

	resources := res.Resources{}
	filenames := []string{}

	bindingName, bindingFilename, svcImageBinding := createSvcImageBinding(cfg, env, p.AppName, p.ServiceName, imageRepo, !isInternalRegistry)
	resources = res.Merge(svcImageBinding, resources)
	filenames = append(filenames, bindingFilename)

	if isInternalRegistry {
		files, regRes, err := imagerepo.CreateInternalRegistryResources(cfg,
			roles.CreateServiceAccount(meta.NamespacedName(cfg.Name, saName)),
			imageRepo, m.GitOpsURL)
		if err != nil {
			return nil, nil, "", fmt.Errorf("failed to get resources for internal image repository: %v", err)
		}
		resources = res.Merge(regRes, resources)
		filenames = append(filenames, files...)
	}

	return filenames, resources, bindingName, nil
}

func createService(serviceName, url string) (*config.Service, error) {
	if url == "" {
		return &config.Service{
			Name: serviceName,
		}, nil
	}
	return &config.Service{
		Name:      serviceName,
		SourceURL: url,
	}, nil
}

func updateKustomization(appFs afero.Fs, base string) error {
	files := res.Resources{}
	filenames, err := environments.ListFiles(appFs, base)
	if err != nil {
		return err
	}
	files[Kustomize] = &res.Kustomization{Resources: filenames.Items()}
	_, err = yaml.WriteResources(appFs, base, files)
	return err
}

func makeSvcImageBindingName(envName, appName, svcName string) string {
	return fmt.Sprintf("%s-%s-%s-binding", envName, appName, svcName)
}

func makeSvcImageBindingFilename(bindingName string) string {
	return filepath.Join("06-bindings", bindingName+".yaml")
}

func makeImageBindingPath(cfg *config.PipelinesConfig, imageRepoBindingFilename string) string {
	return filepath.Join(config.PathForPipelines(cfg), "base", imageRepoBindingFilename)
}

func createSvcImageBinding(cfg *config.PipelinesConfig, env *config.Environment, appName, svcName, imageRepo string, isTLSVerify bool) (string, string, res.Resources) {
	name := makeSvcImageBindingName(env.Name, appName, svcName)
	filename := makeSvcImageBindingFilename(name)
	resourceFilePath := makeImageBindingPath(cfg, filename)
	return name, filename, res.Resources{resourceFilePath: triggers.CreateImageRepoBinding(cfg.Name, name, imageRepo, strconv.FormatBool(isTLSVerify))}
}

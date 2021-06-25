package pipelines

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/mitchellh/go-homedir"
	"github.com/redhat-developer/kam/pkg/pipelines/config"
	"github.com/redhat-developer/kam/pkg/pipelines/environments"
	"github.com/redhat-developer/kam/pkg/pipelines/eventlisteners"
	"github.com/redhat-developer/kam/pkg/pipelines/imagerepo"
	"github.com/redhat-developer/kam/pkg/pipelines/meta"
	res "github.com/redhat-developer/kam/pkg/pipelines/resources"
	"github.com/redhat-developer/kam/pkg/pipelines/roles"
	"github.com/redhat-developer/kam/pkg/pipelines/scm"
	"github.com/redhat-developer/kam/pkg/pipelines/secrets"
	"github.com/redhat-developer/kam/pkg/pipelines/triggers"
	"github.com/redhat-developer/kam/pkg/pipelines/yaml"
	"github.com/spf13/afero"
	corev1 "k8s.io/api/core/v1"
)

// AddServiceOptions control how new services are added to the configuration.
type AddServiceOptions struct {
	AppName             string
	EnvName             string
	GitRepoURL          string
	ImageRepo           string
	PipelinesFolderPath string
	ServiceName         string
	WebhookSecret       string
}

// AddService is the entry-point from the CLI for adding new services.
func AddService(o *AddServiceOptions, appFs afero.Fs) error {
	m, err := config.LoadManifest(appFs, o.PipelinesFolderPath)
	if err != nil {
		return err
	}
	files, otherResources, err := serviceResources(m, appFs, o)
	if err != nil {
		return err
	}

	_, err = yaml.WriteResources(appFs, o.PipelinesFolderPath, files)
	if err != nil {
		return err
	}
	_, err = yaml.WriteResources(appFs, filepath.Join(o.PipelinesFolderPath, ".."), otherResources) // Don't call filepath.ToSlash
	if err != nil {
		return err
	}
	err = createConfigFolder(m, appFs, o)
	if err != nil {
		return fmt.Errorf("Failed to create config folder : %v", err)
	}
	cfg := m.GetPipelinesConfig()
	if cfg != nil {
		base := filepath.ToSlash(filepath.Join(o.PipelinesFolderPath, config.PathForPipelines(cfg), "base"))
		err = updateKustomization(appFs, base)
		if err != nil {
			return err
		}
	}
	return nil
}

func serviceResources(m *config.Manifest, appFs afero.Fs, o *AddServiceOptions) (res.Resources, res.Resources, error) {
	files := res.Resources{}
	otherResources := res.Resources{}
	svc := createService(o.ServiceName, o.GitRepoURL)
	cfg := m.GetPipelinesConfig()
	if cfg != nil && o.WebhookSecret == "" && o.GitRepoURL != "" {
		gitSecret, err := secrets.GenerateString(webhookSecretLength)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to generate service webhook secret: %v", err)
		}
		o.WebhookSecret = gitSecret
	}

	env := m.GetEnvironment(o.EnvName)
	if env == nil {
		return nil, nil, fmt.Errorf("environment %s does not exist", o.EnvName)
	}

	// add the secret only if CI/CD env is present
	if cfg != nil {
		secretName := secrets.MakeServiceWebhookSecretName(o.EnvName, svc.Name)
		var opaqueSecret *corev1.Secret
		var err error
		opaqueSecret, err = secrets.CreateUnsealedSecret(
			meta.NamespacedName(cfg.Name, secretName), o.WebhookSecret,
			eventlisteners.WebhookSecretKey)
		if err != nil {
			return nil, nil, err
		}

		svc.Webhook = &config.Webhook{
			Secret: &config.Secret{
				Name:      secretName,
				Namespace: cfg.Name,
			},
		}
		secretFilename := filepath.ToSlash(filepath.Join("secrets", secretName+".yaml"))
		otherResources[secretFilename] = opaqueSecret

		if m.Config.Pipelines != nil {
			// add the default pipelines if they're absent
			if env.Pipelines == nil {
				repo, err := scm.NewRepository(m.GitOpsURL)
				if err != nil {
					return nil, nil, err
				}
				env.Pipelines = defaultPipelines(repo)
			}

			// use internal registry if no input image registry is provided
			if o.ImageRepo == "" {
				repoName, err := repoFromURL(o.GitRepoURL)
				if err != nil {
					return nil, nil, err
				}
				o.ImageRepo = m.Config.Pipelines.Name + "/" + repoName
			}

			_, resources, bindingName, err := createImageRepoResources(m, cfg, env, o)
			if err != nil {
				return nil, nil, err
			}

			files = res.Merge(resources, files)
			svc.Pipelines = &config.Pipelines{
				Integration: &config.TemplateBinding{
					Bindings: append([]string{bindingName}, env.Pipelines.Integration.Bindings...),
				},
			}
		}
	}

	err := m.AddService(o.EnvName, o.AppName, svc)
	if err != nil {
		return nil, nil, err
	}
	err = m.Validate()
	if err != nil {
		return nil, nil, err
	}

	files[filepath.Base(filepath.Join(o.PipelinesFolderPath, pipelinesFile))] = m // Don't call filepath.ToSlash
	built, err := buildResources(appFs, m)
	if err != nil {
		return nil, nil, err
	}
	return res.Merge(built, files), otherResources, nil
}

func createImageRepoResources(m *config.Manifest, cfg *config.PipelinesConfig, env *config.Environment, p *AddServiceOptions) ([]string, res.Resources, string, error) {
	isInternalRegistry, imageRepo, err := imagerepo.ValidateImageRepo(p.ImageRepo)
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

func createService(serviceName, url string) *config.Service {
	if url == "" {
		return &config.Service{
			Name: serviceName,
		}
	}
	return &config.Service{
		Name:      serviceName,
		SourceURL: url,
	}
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
	return filepath.ToSlash(filepath.Join("06-bindings", bindingName+".yaml"))
}

func makeImageBindingPath(cfg *config.PipelinesConfig, imageRepoBindingFilename string) string {
	return filepath.ToSlash(filepath.Join(config.PathForPipelines(cfg), "base", imageRepoBindingFilename))
}

func createSvcImageBinding(cfg *config.PipelinesConfig, env *config.Environment, appName, svcName, imageRepo string, isTLSVerify bool) (string, string, res.Resources) {
	name := makeSvcImageBindingName(env.Name, appName, svcName)
	filename := makeSvcImageBindingFilename(name)
	resourceFilePath := makeImageBindingPath(cfg, filename)
	return name, filename, res.Resources{resourceFilePath: triggers.CreateImageRepoBinding(cfg.Name, name, imageRepo, strconv.FormatBool(isTLSVerify))}
}

func createConfigFolder(m *config.Manifest, appFs afero.Fs, o *AddServiceOptions) error {
	basePath, err := homedir.Expand(o.PipelinesFolderPath)
	if err != nil {
		return fmt.Errorf("Cannot expand the pipelines.yaml path : %s", o.PipelinesFolderPath)
	}
	env := m.GetEnvironment(o.EnvName)
	app := m.GetApplication(o.EnvName, o.AppName)
	servicePath := config.PathForService(app, env, o.ServiceName)
	finalPath := filepath.Join(basePath, servicePath, "base", "config")
	err = appFs.MkdirAll(finalPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to MkDirAll")
	}
	return nil
}

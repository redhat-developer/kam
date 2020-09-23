package pipelines

import (
	"fmt"

	"github.com/redhat-developer/kam/pkg/pipelines/config"
	res "github.com/redhat-developer/kam/pkg/pipelines/resources"
	"github.com/redhat-developer/kam/pkg/pipelines/scm"
	"github.com/redhat-developer/kam/pkg/pipelines/yaml"
	"github.com/spf13/afero"
)

// EnvParameters encapsulates parameters for add env command.
type EnvParameters struct {
	PipelinesFolderPath string
	EnvName             string
	Cluster             string
}

// AddEnv adds a new environment to the pipelines file.
func AddEnv(o *EnvParameters, appFs afero.Fs) error {
	m, err := config.LoadManifest(appFs, o.PipelinesFolderPath)
	if err != nil {
		return err
	}
	env := m.GetEnvironment(o.EnvName)
	if env != nil {
		return fmt.Errorf("environment %s already exists", o.EnvName)
	}
	files := res.Resources{}
	newEnv, err := newEnvironment(m, o.EnvName)
	if err != nil {
		return err
	}
	if o.Cluster != "" {
		newEnv.Cluster = o.Cluster
	}
	m.Environments = append(m.Environments, newEnv)
	files[pipelinesFile] = m
	buildParams := &BuildParameters{
		PipelinesFolderPath: o.PipelinesFolderPath,
		OutputPath:          o.PipelinesFolderPath,
	}
	built, err := buildResources(appFs, buildParams, m)
	if err != nil {
		return fmt.Errorf("failed to build resources: %v", err)
	}
	files = res.Merge(built, files)
	_, err = yaml.WriteResources(appFs, o.PipelinesFolderPath, files)
	return err
}

func newEnvironment(m *config.Manifest, name string) (*config.Environment, error) {
	pipelinesConfig := m.GetPipelinesConfig()
	if pipelinesConfig != nil && m.GitOpsURL != "" {
		r, err := scm.NewRepository(m.GitOpsURL)
		if err != nil {
			return nil, err
		}
		return &config.Environment{
			Name:      name,
			Pipelines: defaultPipelines(r),
		}, nil
	}

	return &config.Environment{
		Name: name,
	}, nil
}

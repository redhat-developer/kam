package config

import (
	"fmt"

	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/spf13/afero"
)

// LoadManifest reads a manifest file, and configures the environment based on
// the configuration.
func LoadManifest(fs afero.Fs, path string) (*Manifest, error) {
	m, err := ParsePipelinesFolder(fs, path)
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}
	if !(m.Config == nil || m.Config.Git == nil || m.Config.Git.Drivers == nil) {
		drivers := []factory.MappingFunc{}
		for k, v := range m.Config.Git.Drivers {
			drivers = append(drivers, factory.Mapping(k, v))
		}
		if len(drivers) > 0 {
			id := factory.NewDriverIdentifier(drivers...)
			factory.DefaultIdentifier = id
		}
	}
	if err := m.Validate(); err != nil {
		return nil, err
	}
	return m, nil
}

package pipelines

import (
	"fmt"
	"path/filepath"

	"github.com/redhat-developer/gitops-generator/api/v1alpha1"
	gitops "github.com/redhat-developer/gitops-generator/pkg"
	"github.com/redhat-developer/kam/pkg/cmd/component/cmd/ui"
	"github.com/redhat-developer/kam/pkg/pipelines/ioutils"
	"github.com/spf13/afero"
)

// // ComponentOptions is a struct that provides the optional flags
type BootstrapNewOptions struct {
	Output               string //
	ComponentName        string //
	ApplicationName      string //
	Secret               string //
	GitRepoURL           string //
	NameSpace            string //
	TargetPort           int    //
	PushToGit            bool   // If true, gitops repository is pushed to remote git repository.
	Route                string
	Overwrite            bool //
	SaveTokenKeyRing     bool
	PrivateRepoURLDriver string //
}

// Checking if

// function to populate the GitSource struct
func BootstrapNew(o *BootstrapNewOptions, appFs afero.Fs) error {

	componentSpec := v1alpha1.ComponentSpec{
		Application:   o.ApplicationName,
		ComponentName: o.ComponentName,
		Secret:        o.Secret,
		TargetPort:    o.TargetPort,
		Route:         o.Route,

		Source: v1alpha1.ComponentSource{
			ComponentSourceUnion: v1alpha1.ComponentSourceUnion{
				GitSource: &v1alpha1.GitSource{
					URL: o.GitRepoURL,
				},
			},
		},
	}
	BootstrapNewVal := v1alpha1.Component{
		Spec: componentSpec,
	}

	if ui.PathExists(appFs, filepath.Join(o.Output, o.ApplicationName)) && !o.Overwrite {
		return fmt.Errorf("%v the application name already exists in given directory %v", o.ApplicationName, o.Output)
	}
	e := gitops.NewCmdExecutor()
	anyErr := gitops.GenerateAndPush(o.Output, o.GitRepoURL, BootstrapNewVal, e, ioutils.NewFilesystem(), "main", o.PushToGit, "KAM cli", nil)
	if anyErr != nil {
		return fmt.Errorf("failed to create the gitops repository: %q: %w", o.GitRepoURL, anyErr)
	}
	return nil
}

package triggers

import (
	"encoding/json"

	triggersv1 "github.com/tektoncd/triggers/pkg/apis/triggers/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/chetan-rns/gitops-cli/pkg/pipelines/meta"
)

var (
	triggerTemplateTypeMeta = meta.TypeMeta("TriggerTemplate", "triggers.tekton.dev/v1alpha1")
)

const (
	GitRef           = "io.openshift.build.commit.ref"
	GitCommitID      = "io.openshift.build.commit.id"
	GitCommitAuthor  = "io.openshift.build.commit.author"
	GitCommitMessage = "io.openshift.build.commit.message"
	GitCommitDate    = "io.openshift.build.commit.date"
)

// GenerateTemplates will return a slice of trigger templates
func GenerateTemplates(ns, saName string) []triggersv1.TriggerTemplate {
	return []triggersv1.TriggerTemplate{
		CreateDevCDDeployTemplate(ns, saName),
		CreateDevCIBuildPRTemplate(ns, saName),
		CreateCDPushTemplate(ns, saName),
		CreateCIDryRunTemplate(ns, saName),
	}
}

// CreateDevCDDeployTemplate creates DevCDDeployTemplate
func CreateDevCDDeployTemplate(ns, saName string) triggersv1.TriggerTemplate {
	return triggersv1.TriggerTemplate{
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "app-cd-template")),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []triggersv1.ParamSpec{
				createTemplateParamSpec(GitCommitID, "The specific commit SHA."),
				createTemplateParamSpec("gitrepositoryurl", "The git repository url"),
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				{
					RawExtension: runtime.RawExtension{
						Raw: createDevCDResourceTemplate(saName),
					},
				},
			},
		},
	}
}

// CreateDevCIBuildPRTemplate creates DevCIBuildPRTemplate
func CreateDevCIBuildPRTemplate(ns, saName string) triggersv1.TriggerTemplate {
	return triggersv1.TriggerTemplate{
		TypeMeta: triggerTemplateTypeMeta,
		ObjectMeta: meta.ObjectMeta(
			meta.NamespacedName(ns, "app-ci-template")),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []triggersv1.ParamSpec{
				createTemplateParamSpec(GitRef, "The git branch for this PR."),
				createTemplateParamSpec(GitCommitID, "the specific commit SHA."),
				createTemplateParamSpec(GitCommitDate, "The date at which the commit was made"),
				createTemplateParamSpec(GitCommitAuthor, "The name of the github user handle that made the commit"),
				createTemplateParamSpec(GitCommitMessage, "The commit message"),
				createTemplateParamSpec("gitrepositoryurl", "The git repository URL."),
				createTemplateParamSpec("fullname", "The GitHub repository for this PullRequest."),
				createTemplateParamSpec("imageRepo", "The repository to push built images to."),
				createTemplateParamSpec("tlsVerify", "Enable image repostiory TLS certification verification."),
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				{
					RawExtension: runtime.RawExtension{
						Raw: createDevCIResourceTemplate(saName),
					},
				},
			},
		},
	}

}

// CreateCDPushTemplate returns TriggerTemplate for CD Push Request
func CreateCDPushTemplate(ns, saName string) triggersv1.TriggerTemplate {
	return triggersv1.TriggerTemplate{
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "cd-deploy-from-push-template")),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []triggersv1.ParamSpec{

				createTemplateParamSpecDefault(GitRef, "The git revision", "master"),
				createTemplateParamSpec(GitCommitDate, "The date at which the commit was made"),
				createTemplateParamSpec(GitCommitAuthor, "The name of the github user handle that made the commit"),
				createTemplateParamSpec(GitCommitMessage, "The commit message"),
				createTemplateParamSpec("gitrepositoryurl", "The git repository url"),
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				{
					RawExtension: runtime.RawExtension{
						Raw: createCDResourceTemplate(saName),
					},
				},
			},
		},
	}
}

// CreateCIDryRunTemplate returns TriggerTemplate for CI Dry Try
func CreateCIDryRunTemplate(ns, saName string) triggersv1.TriggerTemplate {
	return triggersv1.TriggerTemplate{
		TypeMeta:   triggerTemplateTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "ci-dryrun-from-push-template")),
		Spec: triggersv1.TriggerTemplateSpec{
			Params: []triggersv1.ParamSpec{
				createTemplateParamSpecDefault(GitRef, "The git revision", "master"),
				createTemplateParamSpec(GitCommitID, "The specific commit SHA"),
				createTemplateParamSpec("gitrepositoryurl", "The git repository url"),
			},
			ResourceTemplates: []triggersv1.TriggerResourceTemplate{
				{
					RawExtension: runtime.RawExtension{
						Raw: createCIResourceTemplate(saName),
					},
				},
			},
		},
	}
}

func createTemplateParamSpecDefault(name string, description string, value string) triggersv1.ParamSpec {
	return triggersv1.ParamSpec{
		Name:        name,
		Description: description,
		Default:     strPtr(value),
	}
}

func createTemplateParamSpec(name string, description string) triggersv1.ParamSpec {
	return triggersv1.ParamSpec{
		Name:        name,
		Description: description,
	}
}

func createDevCDResourceTemplate(saName string) []byte {
	byteTemplate, _ := json.Marshal(createDevCDPipelineRun(saName))
	return byteTemplate
}

func createDevCIResourceTemplate(saName string) []byte {
	byteTemplateCI, _ := json.Marshal(createDevCIPipelineRun(saName))
	return byteTemplateCI
}

func createCDResourceTemplate(saName string) []byte {
	byteStageCD, _ := json.Marshal(createCDPipelineRun(saName))
	return byteStageCD
}

func createCIResourceTemplate(saName string) []byte {
	byteStageCI, _ := json.Marshal(createCIPipelineRun(saName))
	return byteStageCI
}

func statusTrackerAnnotations(pipeline, description string) func(*v1.ObjectMeta) {
	return func(om *v1.ObjectMeta) {
		annotations := map[string]string{
			"tekton.dev/git-status":         "true",
			"tekton.dev/status-context":     pipeline,
			"tekton.dev/status-description": description,
		}
		if om.Annotations == nil {
			om.Annotations = map[string]string{}
		}
		for k, v := range annotations {
			om.Annotations[k] = v
		}
	}
}

func strPtr(s string) *string {
	return &s
}

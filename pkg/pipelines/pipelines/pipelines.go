package pipelines

import (
	"fmt"
	"sort"
	"strings"

	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/redhat-developer/kam/pkg/pipelines/meta"
	"github.com/redhat-developer/kam/pkg/pipelines/triggers"
)

var (
	pipelineTypeMeta = meta.TypeMeta("Pipeline", "tekton.dev/v1beta1")
)

const pipelineWorkspace = "shared-data"

// CreateAppCIPipeline creates AppCIPipeline
func CreateAppCIPipeline(name types.NamespacedName) *pipelinev1.Pipeline {
	return &pipelinev1.Pipeline{
		TypeMeta:   pipelineTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Spec: pipelinev1.PipelineSpec{
			Params: paramSpecs(
				"REPO",
				"COMMIT_SHA",
				"TLSVERIFY",
				"BUILD_EXTRA_ARGS",
				"IMAGE",
				"GIT_REF",
				"COMMIT_DATE",
				"COMMIT_AUTHOR",
				"COMMIT_MESSAGE",
				"GIT_REPO"),
			Tasks: []pipelinev1.PipelineTask{
				createGitCloneTask("clone-source"),
				createBuildImageTask("build-image", "clone-source"),
			},
		},
	}
}

func createBuildImageTask(name, runAfter string) pipelinev1.PipelineTask {
	return pipelinev1.PipelineTask{
		Name:     name,
		TaskRef:  createTaskRef("buildah", pipelinev1.ClusterTaskKind),
		RunAfter: []string{runAfter},
		Params: []pipelinev1.Param{
			createTaskParam("TLSVERIFY", "$(params.TLSVERIFY)"),
			createTaskParam("BUILD_EXTRA_ARGS", metadataLabelArgs()),
			createTaskParam("IMAGE", "$(params.IMAGE)"),
		},
	}
}

func createGitCloneTask(name string) pipelinev1.PipelineTask {
	// The output workspace mapping here comes from the git-clone task.
	return pipelinev1.PipelineTask{
		Name:    name,
		TaskRef: createTaskRef("git-clone", pipelinev1.ClusterTaskKind),
		Workspaces: []pipelinev1.WorkspacePipelineTaskBinding{
			{Name: "output", Workspace: pipelineWorkspace},
		},
		Params: []pipelinev1.Param{
			createTaskParam("url", "$(params.GIT_REPO)"),
			createTaskParam("revision", "$(params.GIT_REF)"),
		},
	}
}

// CreateCDPipeline creates a CD pipeline.
func CreateCDPipeline(name types.NamespacedName, stageNamespace string) *pipelinev1.Pipeline {
	return &pipelinev1.Pipeline{
		TypeMeta:   pipelineTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Spec: pipelinev1.PipelineSpec{
			Resources: []pipelinev1.PipelineDeclaredResource{
				createPipelineDeclaredResource("source-repo", "git"),
			},
			Tasks: []pipelinev1.PipelineTask{
				createCDPipelineTask("apply-source"),
			},
		},
	}
}

func createCDPipelineTask(taskName string) pipelinev1.PipelineTask {
	return pipelinev1.PipelineTask{
		Name:    taskName,
		TaskRef: createTaskRef("deploy-from-source-task", pipelinev1.NamespacedTaskKind),
		Resources: &pipelinev1.PipelineTaskResources{
			Inputs: []pipelinev1.PipelineTaskInputResource{createInputTaskResource("source", "source-repo")},
		},
	}
}

// CreateCIPipeline creates CI pipeline
func CreateCIPipeline(name types.NamespacedName, stageNamespace string) *pipelinev1.Pipeline {
	return &pipelinev1.Pipeline{
		TypeMeta:   pipelineTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Spec: pipelinev1.PipelineSpec{

			Resources: []pipelinev1.PipelineDeclaredResource{
				createPipelineDeclaredResource("source-repo", "git"),
			},

			Tasks: []pipelinev1.PipelineTask{
				createCIPipelineTask("apply-source"),
			},
		},
	}
}

// CreateAppCDPipeline creates AppCDPipelin
func CreateAppCDPipeline(name types.NamespacedName, deploymentPath, devNamespace string, isInternalRegistry bool) *pipelinev1.Pipeline {
	return &pipelinev1.Pipeline{
		TypeMeta:   pipelineTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Spec: pipelinev1.PipelineSpec{
			Resources: []pipelinev1.PipelineDeclaredResource{
				createPipelineDeclaredResource("source-repo", "git"),
				createPipelineDeclaredResource("runtime-image", "image"),
			},
			Tasks: []pipelinev1.PipelineTask{
				createDevCDBuildImageTask("build-image"),
				createDevCDDeployImageTask("deploy-image", devNamespace, deploymentPath),
			},
		},
	}
}

func createCIPipelineTask(taskName string) pipelinev1.PipelineTask {
	return pipelinev1.PipelineTask{
		Name:    taskName,
		TaskRef: createTaskRef("deploy-from-source-task", pipelinev1.NamespacedTaskKind),
		Resources: &pipelinev1.PipelineTaskResources{
			Inputs: []pipelinev1.PipelineTaskInputResource{createInputTaskResource("source", "source-repo")},
		},
		Params: []pipelinev1.Param{
			createTaskParam("DRYRUN", "true"),
		},
	}
}

func createDevCDDeployImageTask(name, devNamespace, deploymentPath string) pipelinev1.PipelineTask {
	return pipelinev1.PipelineTask{
		Name:     name,
		TaskRef:  createTaskRef("deploy-using-kubectl-task", pipelinev1.NamespacedTaskKind),
		RunAfter: []string{"build-image"},
		Resources: &pipelinev1.PipelineTaskResources{
			Inputs: []pipelinev1.PipelineTaskInputResource{
				createInputTaskResource("source", "source-repo"),
				createInputTaskResource("image", "runtime-image"),
			},
		},
		Params: []pipelinev1.Param{
			createTaskParam("PATHTODEPLOYMENT", deploymentPath),
			createTaskParam("YAMLPATHTOIMAGE", "spec.template.spec.containers[0].image"),
			createTaskParam("NAMESPACE", devNamespace),
		},
	}
}

func createInputTaskResource(name, resource string) pipelinev1.PipelineTaskInputResource {
	return pipelinev1.PipelineTaskInputResource{
		Name:     name,
		Resource: resource,
	}
}

func createDevCDBuildImageTask(name string) pipelinev1.PipelineTask {
	return pipelinev1.PipelineTask{
		Name:    name,
		TaskRef: createTaskRef("buildah", pipelinev1.ClusterTaskKind),
		Params: []pipelinev1.Param{
			createTaskParam("TLSVERIFY", "true"),
		},
	}
}

func createOutputTaskResource(name, resource string) pipelinev1.PipelineTaskOutputResource {
	return pipelinev1.PipelineTaskOutputResource{
		Name:     name,
		Resource: resource,
	}
}

func createTaskRef(name string, kind pipelinev1.TaskKind) *pipelinev1.TaskRef {
	return &pipelinev1.TaskRef{
		Name: name,
		Kind: kind,
	}
}

func createTaskParam(name, value string) pipelinev1.Param {
	return pipelinev1.Param{
		Name: name,

		Value: pipelinev1.ArrayOrString{
			Type:      pipelinev1.ParamTypeString,
			StringVal: value,
		},
	}
}

func createPipelineDeclaredResource(name, resourceType string) pipelinev1.PipelineDeclaredResource {
	return pipelinev1.PipelineDeclaredResource{Name: name, Type: resourceType}
}

func metadataLabelArgs() string {
	labels := map[string]string{
		triggers.GitCommitID:      "$(params.COMMIT_SHA)",
		triggers.GitRef:           "$(params.GIT_REF)",
		triggers.GitCommitDate:    "$(params.COMMIT_DATE)",
		triggers.GitCommitAuthor:  "$(params.COMMIT_AUTHOR)",
		triggers.GitCommitMessage: "$(params.COMMIT_MESSAGE)",
	}
	labelArgs := []string{}
	for k, v := range labels {
		labelArgs = append(labelArgs, fmt.Sprintf("--label=%s='%s'", k, v))
	}
	sort.Strings(labelArgs)
	return strings.Join(labelArgs, " ")
}

func paramSpecs(s ...string) []pipelinev1.ParamSpec {
	specs := make([]pipelinev1.ParamSpec, len(s))
	for i := range s {
		specs[i] = paramSpec(s[i])
	}
	return specs
}

func paramSpec(name string) pipelinev1.ParamSpec {
	return pipelinev1.ParamSpec{Name: name, Type: "string"}
}

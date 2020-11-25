package triggers

import (
	pipelinev1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/redhat-developer/kam/pkg/pipelines/meta"
)

var (
	pipelineRunTypeMeta = meta.TypeMeta("PipelineRun", "tekton.dev/v1beta1")
)

func createDevCDPipelineRun(saName string) pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "app-cd-pipeline-run-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: saName,
			PipelineRef:        createPipelineRef("app-cd-pipeline"),
			Resources:          createDevResource("$(params." + GitCommitID + ")"),
		},
	}
}

func createDevCIPipelineRun(saName string) pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "app-ci-pipeline-run-$(uid)"), statusTrackerAnnotations("dev-ci-build-from-pr", "CI build on push event")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: saName,
			PipelineRef:        createPipelineRef("app-ci-pipeline"),
			Params: []pipelinev1.Param{
				createPipelineBindingParam("REPO", "$(params.fullname)"),
				createPipelineBindingParam("GIT_REPO", "$(params.gitrepositoryurl)"),
				createPipelineBindingParam("TLSVERIFY", "$(params.tlsVerify)"),
				createPipelineBindingParam("BUILD_EXTRA_ARGS", "$(params.build_extra_args)"),
				createPipelineBindingParam("IMAGE", "$(params.imageRepo):$(params."+GitRef+")-$(params."+GitCommitID+")"),
				createPipelineBindingParam("COMMIT_SHA", "$(params."+GitCommitID+")"),
				createPipelineBindingParam("GIT_REF", "$(params."+GitRef+")"),
				createPipelineBindingParam("COMMIT_DATE", "$(params."+GitCommitDate+")"),
				createPipelineBindingParam("COMMIT_AUTHOR", "$(params."+GitCommitAuthor+")"),
				createPipelineBindingParam("COMMIT_MESSAGE", "$(params."+GitCommitMessage+")"),
			},
			Resources: createDevResource("$(params." + GitCommitID + ")"),
		},
	}
}

// {
// 	Name: "runtime-image",
// 	ResourceSpec: &pipelinev1alpha1.PipelineResourceSpec{
// 		Type: "image",
// 		Params: []pipelinev1.ResourceParam{
// 			createResourceParams("url", "$(params.imageRepo):$(params."+GitRef+")-$(params."+GitCommitID+")"),
// 		},
// 	},
// },

func createCDPipelineRun(saName string) pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "cd-deploy-from-push-pipeline-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: saName,
			PipelineRef:        createPipelineRef("cd-deploy-from-push-pipeline"),
			Resources:          createResources(),
		},
	}
}

func createCIPipelineRun(saName string) pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "ci-dryrun-from-push-pipeline-$(uid)"), statusTrackerAnnotations("ci-dryrun-from-push-pipeline", "CI dry run on push event")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: saName,
			PipelineRef:        createPipelineRef("ci-dryrun-from-push-pipeline"),
			Resources:          createResources(),
		},
	}
}

func createDevResource(revision string) []pipelinev1.PipelineResourceBinding {
	return []pipelinev1.PipelineResourceBinding{
		{
			Name: "source-repo",
			ResourceSpec: &pipelinev1alpha1.PipelineResourceSpec{
				Type: "git",
				Params: []pipelinev1.ResourceParam{
					createResourceParams("revision", revision),
					createResourceParams("url", "$(params.gitrepositoryurl)"),
				},
			},
		},
	}
}

func createResources() []pipelinev1.PipelineResourceBinding {
	return []pipelinev1.PipelineResourceBinding{
		{
			Name: "source-repo",
			ResourceSpec: &pipelinev1alpha1.PipelineResourceSpec{
				Type: "git",
				Params: []pipelinev1.ResourceParam{
					createResourceParams("revision", "$(params."+GitCommitID+")"),
					createResourceParams("url", "$(params.gitrepositoryurl)"),
				},
			},
		},
	}
}

func createResourceParams(name, value string) pipelinev1.ResourceParam {
	return pipelinev1.ResourceParam{
		Name:  name,
		Value: value,
	}
}
func createPipelineRef(name string) *pipelinev1.PipelineRef {
	return &pipelinev1.PipelineRef{
		Name: name,
	}
}

func createPipelineBindingParam(name, value string) pipelinev1.Param {
	return pipelinev1.Param{
		Name: name,
		Value: pipelinev1.ArrayOrString{
			StringVal: value,
			Type:      pipelinev1.ParamTypeString,
		},
	}
}

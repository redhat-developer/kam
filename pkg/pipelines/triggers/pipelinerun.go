package triggers

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	pipelineresource "github.com/tektoncd/pipeline/pkg/apis/resource/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

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
			Resources:          createDevResource("$(tt.params." + GitCommitID + ")"),
		},
	}
}

func createDevCIPipelineRun(saName string) pipelinev1.PipelineRun {
	return pipelinev1.PipelineRun{
		TypeMeta: pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(
			meta.NamespacedName("", "app-ci-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: saName,
			PipelineRef:        createPipelineRef("app-ci-pipeline"),
			Params: []pipelinev1.Param{
				createPipelineBindingParam("REPO", "$(tt.params.fullname)"),
				createPipelineBindingParam("GIT_REPO", "$(tt.params.gitrepositoryurl)"),
				createPipelineBindingParam("TLSVERIFY", "$(tt.params.tlsVerify)"),
				createPipelineBindingParam("BUILD_EXTRA_ARGS", "$(tt.params.build_extra_args)"),
				createPipelineBindingParam("IMAGE", "$(tt.params.imageRepo):$(tt.params."+GitRef+")-$(tt.params."+GitCommitID+")"),
				createPipelineBindingParam("COMMIT_SHA", "$(tt.params."+GitCommitID+")"),
				createPipelineBindingParam("GIT_REF", "$(tt.params."+GitRef+")"),
				createPipelineBindingParam("COMMIT_DATE", "$(tt.params."+GitCommitDate+")"),
				createPipelineBindingParam("COMMIT_AUTHOR", "$(tt.params."+GitCommitAuthor+")"),
				createPipelineBindingParam("COMMIT_MESSAGE", "$(tt.params."+GitCommitMessage+")"),
			},
			Workspaces: []pipelinev1.WorkspaceBinding{
				{
					Name: "shared-data",
					VolumeClaimTemplate: &corev1.PersistentVolumeClaim{
						Spec: corev1.PersistentVolumeClaimSpec{
							AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{"storage": resource.MustParse("1Gi")},
							},
						},
					},
				},
			},
		},
	}
}

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
		TypeMeta: pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(
			meta.NamespacedName("", "ci-dryrun-from-push-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: saName,
			PipelineRef:        createPipelineRef("ci-dryrun-from-push-pipeline"),
			Resources:          createResources(),
			Params: []pipelinev1.Param{
				createPipelineBindingParam("REPO", "$(tt.params.fullname)"),
				createPipelineBindingParam("GIT_REPO", "$(tt.params.gitrepositoryurl)"),
				createPipelineBindingParam("COMMIT_SHA", "$(tt.params.io.openshift.build.commit.id)"),
			},
		},
	}
}

func createDevResource(revision string) []pipelinev1.PipelineResourceBinding {
	return []pipelinev1.PipelineResourceBinding{
		{
			Name: "source-repo",
			ResourceSpec: &pipelineresource.PipelineResourceSpec{
				Type: "git",
				Params: []pipelinev1.ResourceParam{
					createResourceParams("revision", revision),
					createResourceParams("url", "$(tt.params.gitrepositoryurl)"),
				},
			},
		},
	}
}

func createResources() []pipelinev1.PipelineResourceBinding {
	return []pipelinev1.PipelineResourceBinding{
		{
			Name: "source-repo",
			ResourceSpec: &pipelineresource.PipelineResourceSpec{
				Type: "git",
				Params: []pipelinev1.ResourceParam{
					createResourceParams("revision", "$(tt.params."+GitCommitID+")"),
					createResourceParams("url", "$(tt.params.gitrepositoryurl)"),
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

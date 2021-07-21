package pipelines

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/redhat-developer/kam/pkg/pipelines/meta"
)

func Test_createTaskParam(t *testing.T) {
	p := createTaskParam("TEST_PARAM", "$(params.test_param)")

	want := pipelinev1.Param{
		Name: "TEST_PARAM",
		Value: pipelinev1.ArrayOrString{
			Type:      "string",
			StringVal: "$(params.test_param)",
		},
	}

	if diff := cmp.Diff(want, p); diff != "" {
		t.Fatalf("createTaskParam failed:\n%s", diff)
	}
}

func Test_paramSpec(t *testing.T) {
	ps := paramSpec("testing")
	want := pipelinev1.ParamSpec{Name: "testing", Type: "string"}

	if diff := cmp.Diff(want, ps); diff != "" {
		t.Fatalf("paramSpec failed:\n%s", diff)
	}
}

func Test_paramSpecs(t *testing.T) {
	ps := paramSpecs("testing1", "testing2")
	want := []pipelinev1.ParamSpec{
		{Name: "testing1", Type: "string"},
		{Name: "testing2", Type: "string"},
	}

	if diff := cmp.Diff(want, ps); diff != "" {
		t.Fatalf("paramSpec failed:\n%s", diff)
	}

}

func TestCreateAppCIPipeline(t *testing.T) {
	name := types.NamespacedName{Name: "test-pipeline", Namespace: "test-ns"}
	p := CreateAppCIPipeline(name)

	want := &pipelinev1.Pipeline{
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
			Workspaces: []pipelinev1.PipelineWorkspaceDeclaration{
				{Name: pipelineWorkspace, Description: "This workspace will receive the cloned git repo."},
			},
			Tasks: []pipelinev1.PipelineTask{
				createCommitStatusPipelineTask(PendingCommitStatusTask, "pending", "The build has started"),

				{
					Name:    "clone-source",
					TaskRef: &pipelinev1.TaskRef{Name: "git-clone", Kind: "ClusterTask"},
					Params: []pipelinev1.Param{
						createTaskParam("url", "$(params.GIT_REPO)"),
						createTaskParam("revision", "$(params.GIT_REF)"),
					},
					Workspaces: []pipelinev1.WorkspacePipelineTaskBinding{
						{Name: "output", Workspace: pipelineWorkspace},
					},
					RunAfter: []string{PendingCommitStatusTask},
				},

				{
					Name:     "build-image",
					RunAfter: []string{"clone-source"},
					TaskRef:  &pipelinev1.TaskRef{Name: "buildah", Kind: "ClusterTask"},
					Workspaces: []pipelinev1.WorkspacePipelineTaskBinding{
						{Name: "source", Workspace: pipelineWorkspace},
					},
					Params: []pipelinev1.Param{
						createTaskParam("TLSVERIFY", "$(params.TLSVERIFY)"),
						{
							Name: "BUILD_EXTRA_ARGS",
							Value: pipelinev1.ArrayOrString{
								Type:      "string",
								StringVal: metadataLabelArgs(),
							},
						},
						createTaskParam("IMAGE", "$(params.IMAGE)"),
					},
				},
			},
			Finally: []v1beta1.PipelineTask{
				createCommitStatusPipelineTask("set-final-status", "$(tasks.build-image.status)", "The build is complete"),
			},
		},
	}

	if diff := cmp.Diff(want, p); diff != "" {
		t.Fatalf("CreateAppCIPipeline failed:\n%s", diff)
	}
}

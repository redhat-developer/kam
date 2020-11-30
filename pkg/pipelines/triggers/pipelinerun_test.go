package triggers

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/redhat-developer/kam/pkg/pipelines/meta"
)

var (
	sName = "pipeline"
)

func TestCreateDevCDPipelineRun(t *testing.T) {
	validDevCDPipeline := pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "app-cd-pipeline-run-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: sName,
			PipelineRef:        createPipelineRef("app-cd-pipeline"),
			Resources:          createDevResource("$(params.io.openshift.build.commit.id)"),
		},
	}
	template := createDevCDPipelineRun(sName)
	if diff := cmp.Diff(validDevCDPipeline, template); diff != "" {
		t.Fatalf("createDevCDPipelineRun failed:\n%s", diff)
	}
}

func TestCreateDevCIPipelineRun(t *testing.T) {
	want := pipelinev1.PipelineRun{
		TypeMeta: pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(
			meta.NamespacedName("", "app-ci-pipeline-run-$(uid)"),
			func(om *metav1.ObjectMeta) {
				om.Annotations = map[string]string{
					"tekton.dev/commit-status-source-sha": "$(params.io.openshift.build.commit.id)",
					"tekton.dev/commit-status-source-url": "$(params.gitrepositoryurl)",
					"tekton.dev/git-status":               "true",
					"tekton.dev/status-context":           "dev-ci-build-from-pr",
					"tekton.dev/status-description":       "CI build on push event",
				}
			}),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: sName,
			PipelineRef:        createPipelineRef("app-ci-pipeline"),
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
			Params: []pipelinev1.Param{
				createPipelineBindingParam("REPO", "$(params.fullname)"),
				createPipelineBindingParam("GIT_REPO", "$(params.gitrepositoryurl)"),
				createPipelineBindingParam("TLSVERIFY", "$(params.tlsVerify)"),
				createPipelineBindingParam("BUILD_EXTRA_ARGS", "$(params.build_extra_args)"),
				createPipelineBindingParam("IMAGE", "$(params.imageRepo):$(params."+GitRef+")-$(params."+GitCommitID+")"),
				createPipelineBindingParam("COMMIT_SHA", "$(params.io.openshift.build.commit.id)"),
				createPipelineBindingParam("GIT_REF", "$(params.io.openshift.build.commit.ref)"),
				createPipelineBindingParam("COMMIT_DATE", "$(params.io.openshift.build.commit.date)"),
				createPipelineBindingParam("COMMIT_AUTHOR", "$(params.io.openshift.build.commit.author)"),
				createPipelineBindingParam("COMMIT_MESSAGE", "$(params.io.openshift.build.commit.message)"),
			},
		},
	}
	template := createDevCIPipelineRun(sName)
	if diff := cmp.Diff(want, template); diff != "" {
		t.Fatalf("createDevCIPipelineRun failed:\n%s", diff)
	}
}

func TestCreateCDPipelineRun(t *testing.T) {
	want := pipelinev1.PipelineRun{
		TypeMeta:   pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("", "cd-deploy-from-push-pipeline-$(uid)")),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: sName,
			PipelineRef:        createPipelineRef("cd-deploy-from-push-pipeline"),
			Resources:          createResources(),
		},
	}
	template := createCDPipelineRun(sName)
	if diff := cmp.Diff(want, template); diff != "" {
		t.Fatalf("createCDPipelineRun failed:\n%s", diff)
	}
}

func TestCreateStageCIPipelineRun(t *testing.T) {
	want := pipelinev1.PipelineRun{
		TypeMeta: pipelineRunTypeMeta,
		ObjectMeta: meta.ObjectMeta(
			meta.NamespacedName("", "ci-dryrun-from-push-pipeline-$(uid)"),
			statusTrackerAnnotations("ci-dryrun-from-push-pipeline", "CI dry run on push event", nil)),
		Spec: pipelinev1.PipelineRunSpec{
			ServiceAccountName: sName,
			PipelineRef:        createPipelineRef("ci-dryrun-from-push-pipeline"),
			Resources:          createResources(),
		},
	}
	template := createCIPipelineRun(sName)
	if diff := cmp.Diff(want, template); diff != "" {
		t.Fatalf("createCIPipelineRun failed:\n%s", diff)
	}
}

func TestCreateDevResource(t *testing.T) {
	want := []pipelinev1.PipelineResourceBinding{
		{
			Name: "source-repo",
			ResourceSpec: &pipelinev1alpha1.PipelineResourceSpec{
				Type: "git",
				Params: []pipelinev1.ResourceParam{
					createResourceParams("revision", "test"),
					createResourceParams("url", "$(params.gitrepositoryurl)"),
				},
			},
		},
	}
	got := createDevResource("test")
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("createDevResource() failed: \n%s", diff)
	}
}

package tasks

import (
	"github.com/redhat-developer/kam/pkg/pipelines/meta"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

// CreateCommitStatusTask creates a task to add commit status
func CreateCommitStatusTask(namespace string) *pipelinev1.Task {
	return &pipelinev1.Task{
		TypeMeta:   taskTypeMeta,
		ObjectMeta: meta.ObjectMeta(types.NamespacedName{Name: "set-commit-status", Namespace: namespace}),
		Spec: pipelinev1.TaskSpec{
			Params: []v1beta1.ParamSpec{
				createTaskParam("GIT_REPO", "", pipelinev1.ParamTypeString),
				createTaskParam("REPO", "", pipelinev1.ParamTypeString),
				createTaskParamWithDefault("GIT_TOKEN_SECRET_NAME", "", pipelinev1.ParamTypeString, "git-host-access-token"),
				createTaskParamWithDefault("GIT_TOKEN_SECRET_KEY", "", pipelinev1.ParamTypeString, "token"),
				createTaskParam("COMMIT_SHA", "", pipelinev1.ParamTypeString),
				createTaskParam("DESCRIPTION", "", pipelinev1.ParamTypeString),
				createTaskParamWithDefault("CONTEXT", "", pipelinev1.ParamTypeString, "continous-integration/tekton"),
				createTaskParam("STATE", "", pipelinev1.ParamTypeString),
			},
			Steps: []v1beta1.Step{
				{
					Container: v1.Container{
						Name:  "set-commit-status",
						Image: "quay.io/redhat-developer/gitops-commit-status@sha256:b28e5d207f70fc37f5cbf79fabcfa2708459316d09e9def9647c06807d9464cd",
						Env: []v1.EnvVar{
							{
								Name: "GITHOSTACCESSTOKEN",
								ValueFrom: &v1.EnvVarSource{
									SecretKeyRef: &v1.SecretKeySelector{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "$(params.GIT_TOKEN_SECRET_NAME)",
										},
										Key: "$(params.GIT_TOKEN_SECRET_KEY)",
									},
								},
							},
						},
					},
					Script: "gitops-commit-status --url $(params.GIT_REPO) --path $(params.REPO) --sha $(params.COMMIT_SHA) --context $(params.CONTEXT) --status $(params.STATE)",
				},
			},
		},
	}
}

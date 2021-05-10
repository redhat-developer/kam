package tasks

import (
	"github.com/redhat-developer/kam/pkg/pipelines/meta"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

const commitStatusScript = `#!/usr/libexec/platform-python
import json
import os
import http.client
status_url = "$(params.API_PATH_PREFIX)" + "/repos/$(params.REPO)/" + \
	"statuses/$(params.COMMIT_SHA)"
data = {
	"state": "$(params.STATE)",
	"description": "$(params.DESCRIPTION)",
	"context": "$(params.CONTEXT)"
}
conn = http.client.HTTPSConnection("$(params.GIT_REPO)")
r = conn.request(
	"POST",
	status_url,
	body=json.dumps(data),
	headers={
		"User-Agent": "TektonCD, the peaceful cat",
		"Authorization": "Bearer " + os.environ["GITHUBTOKEN"],
		"Accept": "application/vnd.github.v3+json ",
	})
resp = conn.getresponse()
if not str(resp.status).startswith("2"):
	print("Error: %d" % (resp.status))
	print(resp.read())
else:
  	print("GitHub status '$(params.STATE)' has been set on " "$(params.REPO)#$(params.COMMIT_SHA) ")`

// CreateCommitStatusTask creates a task to add commit status
func CreateCommitStatusTask(namespace string) *pipelinev1.Task {
	return &pipelinev1.Task{
		TypeMeta:   taskTypeMeta,
		ObjectMeta: meta.ObjectMeta(types.NamespacedName{Name: "set-commit-status", Namespace: namespace}),
		Spec: pipelinev1.TaskSpec{
			Params: []v1beta1.ParamSpec{
				createTaskParamWithDefault("GIT_REPO", "", pipelinev1.ParamTypeString, "api.github.com"),
				createTaskParamWithDefault("API_PATH_PREFIX", "", pipelinev1.ParamTypeString, ""),
				createTaskParam("REPO", "", pipelinev1.ParamTypeString),
				createTaskParamWithDefault("GITHUB_TOKEN_SECRET_NAME", "", pipelinev1.ParamTypeString, "git-host-access-token"),
				createTaskParamWithDefault("GITHUB_TOKEN_SECRET_KEY", "", pipelinev1.ParamTypeString, "token"),
				createTaskParam("COMMIT_SHA", "", pipelinev1.ParamTypeString),
				createTaskParam("DESCRIPTION", "", pipelinev1.ParamTypeString),
				createTaskParamWithDefault("CONTEXT", "", pipelinev1.ParamTypeString, "continous-integration/tekton"),
				createTaskParam("STATE", "", pipelinev1.ParamTypeString),
			},
			Steps: []v1beta1.Step{
				{
					Container: v1.Container{
						Name:  "set-commit-status",
						Image: "registry.access.redhat.com/ubi8/python-38:1-34.1599745032",
						Env: []v1.EnvVar{
							{
								Name: "GITHUBTOKEN",
								ValueFrom: &v1.EnvVarSource{
									SecretKeyRef: &v1.SecretKeySelector{
										LocalObjectReference: v1.LocalObjectReference{
											Name: "$(params.GITHUB_TOKEN_SECRET_NAME)",
										},
										Key: "$(params.GITHUB_TOKEN_SECRET_KEY)",
									},
								},
							},
						},
					},
					Script: commitStatusScript,
				},
			},
		},
	}
}

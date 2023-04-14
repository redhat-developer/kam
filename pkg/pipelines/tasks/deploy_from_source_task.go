package tasks

import (
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"github.com/redhat-developer/kam/pkg/pipelines/meta"
)

// CreateDeployFromSourceTask creates DeployFromSourceTask
func CreateDeployFromSourceTask(ns, script string) pipelinev1.Task {
	task := pipelinev1.Task{
		TypeMeta:   taskTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, "deploy-from-source-task")),
		Spec: pipelinev1.TaskSpec{
			Params:    paramsForDeploymentFromSourceTask(),
			Resources: createResourcesForDeployFromSourceTask(),
			Steps:     createStepsForDeployFromSourceTask(script),
		},
	}
	return task
}

func createStepsForDeployFromSourceTask(script string) []pipelinev1.Step {
	return []pipelinev1.Step{
		{
			Name:       "run-kubectl",
			Image:      "quay.io/redhat-developer/k8s-kubectl",
			WorkingDir: "/workspace/source",
			Command:    nil,
			Args:       nil,
			Script:     script,
		},
	}
}

func paramsForDeploymentFromSourceTask() []pipelinev1.ParamSpec {
	return []pipelinev1.ParamSpec{
		createTaskParamWithDefault(
			"DRYRUN",
			"If true run a server-side dryrun.",
			pipelinev1.ParamTypeString,
			"false",
		),
	}
}

func createResourcesForDeployFromSourceTask() *pipelinev1.TaskResources {
	return &pipelinev1.TaskResources{
		Inputs: []pipelinev1.TaskResource{
			createTaskResource("source", "git"),
		},
	}
}

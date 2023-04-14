package tasks

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	pipelinev1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testNS = "testing-ns"

func TestDeployFromSourceTask(t *testing.T) {
	wantedTask := pipelinev1.Task{
		TypeMeta: taskTypeMeta,
		ObjectMeta: v1.ObjectMeta{
			Name:      "deploy-from-source-task",
			Namespace: testNS,
		},
		Spec: pipelinev1.TaskSpec{
			Params:    paramsForDeploymentFromSourceTask(),
			Resources: createResourcesForDeployFromSourceTask(),
			Steps: []pipelinev1.Step{
				{
					Name:       "run-kubectl",
					Image:      "quay.io/redhat-developer/k8s-kubectl",
					WorkingDir: "/workspace/source",
					Script:     "test",
				},
			},
		},
	}
	deployFromSourceTask := CreateDeployFromSourceTask(testNS, "test")
	if diff := cmp.Diff(wantedTask, deployFromSourceTask); diff != "" {
		t.Fatalf("CreateDeployFromSourceTask() failed \n%s", diff)
	}
}

func TestCreateTaskParamWithDefault(t *testing.T) {
	validTaskParam := pipelinev1.ParamSpec{
		Name:        "sample",
		Type:        pipelinev1.ParamTypeString,
		Description: "sample",
		Default: &pipelinev1.ArrayOrString{
			StringVal: "sample",
			Type:      "string",
		},
	}
	taskParam := createTaskParamWithDefault("sample", "sample", pipelinev1.ParamTypeString, "sample")
	if diff := cmp.Diff(validTaskParam, taskParam); diff != "" {
		t.Fatalf("createTaskParamWithDefault() failed:\n%s", diff)
	}
}

func TestCreateTaskParam(t *testing.T) {
	validTaskParam := pipelinev1.ParamSpec{
		Name:        "sample",
		Type:        pipelinev1.ParamTypeString,
		Description: "sample",
	}
	taskParam := createTaskParam("sample", "sample", pipelinev1.ParamTypeString)
	if diff := cmp.Diff(validTaskParam, taskParam); diff != "" {
		t.Fatalf("createTaskParam() failed:\n%s", diff)
	}
}

func TestCreateContainer(t *testing.T) {
	validContainer := corev1.Container{
		Name:       "sampleName",
		Image:      "sampleImage",
		WorkingDir: "sampleDir",
		Command:    []string{"sample"},
		Args:       []string{"sample"},
	}
	container := createContainer("sampleName", "sampleImage", "sampleDir", []string{"sample"}, []string{"sample"})
	if diff := cmp.Diff(validContainer, container); diff != "" {
		t.Fatalf("createContainer() failed:\n%s", diff)
	}
}

func TestCreateTaskResource(t *testing.T) {
	validTaskResource := pipelinev1.TaskResource{
		ResourceDeclaration: pipelinev1.ResourceDeclaration{
			Name: "sample",
			Type: "git",
		},
	}
	taskResource := createTaskResource("sample", "git")
	if diff := cmp.Diff(validTaskResource, taskResource); diff != "" {
		t.Fatalf("createTaskResource() failed:\n%s", diff)
	}
}

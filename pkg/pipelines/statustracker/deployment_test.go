package statustracker

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/google/go-cmp/cmp"
	"github.com/redhat-developer/gitops-cli/pkg/pipelines/deployment"
	"github.com/redhat-developer/gitops-cli/pkg/pipelines/meta"
	res "github.com/redhat-developer/gitops-cli/pkg/pipelines/resources"
	"github.com/redhat-developer/gitops-cli/pkg/pipelines/roles"
)

const testRepoURL = "https://github.com/testing/testing.git"

func TestCreateStatusTrackerDeployment(t *testing.T) {
	deploy := createStatusTrackerDeployment("dana-cicd", testRepoURL, "")
	want := &appsv1.Deployment{
		TypeMeta: meta.TypeMeta("Deployment", "apps/v1"),
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName("dana-cicd", operatorName), meta.AddLabels(
			map[string]string{
				deployment.KubernetesAppNameLabel: operatorName,
				deployment.KubernetesPartOfLabel:  commitStatusAppLabel,
			},
		)),
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr32(1),
			Selector: deployment.LabelSelector(operatorName, commitStatusAppLabel),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						deployment.KubernetesAppNameLabel: operatorName,
						deployment.KubernetesPartOfLabel:  commitStatusAppLabel,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: operatorName,
					Containers: []corev1.Container{
						{
							Name:            operatorName,
							Image:           containerImage,
							Command:         []string{operatorName},
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name: "WATCH_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
								},
								{
									Name: "POD_NAME",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.name",
										},
									},
								},
								{
									Name:  "OPERATOR_NAME",
									Value: operatorName,
								},
							},
						},
					},
				},
			},
		},
	}

	if diff := cmp.Diff(want, deploy); diff != "" {
		t.Fatalf("deployment diff: %s", diff)
	}
}

func TestResource(t *testing.T) {

	ns := "my-test-ns"
	generated, err := Resources(ns, testRepoURL, "")
	if err != nil {
		t.Fatal(err)
	}
	name := meta.NamespacedName(ns, operatorName)
	sa := roles.CreateServiceAccount(name)
	want := res.Resources{
		"02-rolebindings/commit-status-tracker-service-account.yaml": sa,
		"02-rolebindings/commit-status-tracker-role.yaml":            roles.CreateRole(name, roleRules),
		"02-rolebindings/commit-status-tracker-rolebinding.yaml":     roles.CreateRoleBinding(name, sa, "Role", operatorName),
		"10-commit-status-tracker/operator.yaml":                     createStatusTrackerDeployment(ns, "https://github.com/testing/testing.git", ""),
	}

	if diff := cmp.Diff(want, generated); diff != "" {
		t.Fatalf("deployment diff: %s", diff)
	}
}

func TestMakeEnvironmentWithCustomDriver(t *testing.T) {
	want := []corev1.EnvVar{
		{
			Name: "WATCH_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name:  "OPERATOR_NAME",
			Value: operatorName,
		},
		{
			Name:  "GIT_DRIVERS",
			Value: "gitlab.example.com=gitlab",
		},
	}
	customRepoURL := "https://gitlab.example.com"
	if diff := cmp.Diff(want, makeEnvironment(customRepoURL, "gitlab")); diff != "" {
		t.Fatalf("custom environment:\n%s", diff)
	}
}

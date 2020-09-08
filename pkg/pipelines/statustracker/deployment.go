package statustracker

import (
	"fmt"
	"net/url"

	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealed-secrets/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/deployment"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/meta"
	res "github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/resources"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/roles"
	"github.com/rhd-gitops-example/gitops-cli/pkg/pipelines/secrets"
)

const (
	operatorName         = "commit-status-tracker"
	containerImage       = "quay.io/redhat-developer/commit-status-tracker:v0.0.3"
	commitStatusAppLabel = "commit-status-tracker-operator"
)

type secretSealer = func(types.NamespacedName, types.NamespacedName, string, string) (*ssv1alpha1.SealedSecret, error)

var defaultSecretSealer secretSealer = secrets.CreateSealedSecret

var (
	roleRules = []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"pods", "services", "services/finalizers", "endpoints", "persistentvolumeclaims", "events", "configmaps", "secrets"},
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
		},
		{
			APIGroups: []string{"apps"},
			Resources: []string{"deployments", "daemonsets", "replicasets", "statefulsets"},
			Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
		},
		{
			APIGroups: []string{"monitoring.coreos.com"},
			Resources: []string{"servicemonitors"},
			Verbs:     []string{"get", "create"},
		},
		{
			APIGroups:     []string{"apps"},
			Resources:     []string{"deployments/finalizers"},
			ResourceNames: []string{operatorName},
			Verbs:         []string{"update"},
		},
		{
			APIGroups: []string{""},
			Resources: []string{"pods"},
			Verbs:     []string{"get"},
		},
		{
			APIGroups: []string{"apps"},
			Resources: []string{"replicasets", "deployments"},
			Verbs:     []string{"get"},
		},
		{
			APIGroups: []string{"tekton.dev"},
			Resources: []string{"pipelineruns"},
			Verbs:     []string{"get", "list", "watch"},
		},
	}
)

func createStatusTrackerDeployment(ns, repoURL, driver string) *appsv1.Deployment {
	return deployment.Create(commitStatusAppLabel, ns, operatorName, containerImage,
		deployment.ServiceAccount(operatorName),
		deployment.Env(makeEnvironment(repoURL, driver)),
		deployment.Command([]string{operatorName}))
}

// Resources returns a list of newly created resources that are required start
// the status-tracker service.
func Resources(ns, token string, sealedSecretsservice types.NamespacedName, repoURL, driver string) (res.Resources, error) {
	name := meta.NamespacedName(ns, operatorName)
	sa := roles.CreateServiceAccount(name)

	githubAuth, err := defaultSecretSealer(meta.NamespacedName(ns, "commit-status-tracker-git-secret"), sealedSecretsservice, token, "token")
	if err != nil {
		return nil, fmt.Errorf("failed to generate Status Tracker Secret: %v", err)
	}
	return res.Resources{
		"02-rolebindings/commit-status-tracker-role.yaml":            roles.CreateRole(name, roleRules),
		"02-rolebindings/commit-status-tracker-rolebinding.yaml":     roles.CreateRoleBinding(name, sa, "Role", operatorName),
		"02-rolebindings/commit-status-tracker-service-account.yaml": sa,
		"03-secrets/commit-status-tracker.yaml":                      githubAuth,
		"10-commit-status-tracker/operator.yaml":                     createStatusTrackerDeployment(ns, repoURL, driver),
	}, nil
}

func ptr32(i int32) *int32 {
	return &i
}

func makeEnvironment(repoURL, driver string) []corev1.EnvVar {
	vars := []corev1.EnvVar{
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
	}
	if host := hostFromURL(repoURL); driver != "" && host != "" {
		vars = append(vars, corev1.EnvVar{
			Name:  "GIT_DRIVERS",
			Value: fmt.Sprintf("%s=%s", host, driver),
		})
	}
	return vars
}

func hostFromURL(s string) string {
	p, err := url.Parse(s)
	if err != nil {
		return ""
	}
	return p.Host
}

package statustracker

import (
	"fmt"
	"net/url"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/redhat-developer/kam/pkg/pipelines/deployment"
	"github.com/redhat-developer/kam/pkg/pipelines/meta"
	res "github.com/redhat-developer/kam/pkg/pipelines/resources"
	"github.com/redhat-developer/kam/pkg/pipelines/roles"
)

const (
	operatorName         = "commit-status-tracker"
	containerImage       = "quay.io/redhat-developer/commit-status-tracker:v0.0.4"
	commitStatusAppLabel = "commit-status-tracker-operator"

	// CommitStatusTrackerSecret is used by commit-status-tracker to
	// authenticate Git requests.
	CommitStatusTrackerSecret = "git-host-access-token"
)

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

// Resources returns a list of newly created resources that are required to
// setup the status-tracker service.
func Resources(ns, repoURL, driver string) (res.Resources, error) {
	name := meta.NamespacedName(ns, operatorName)
	sa := roles.CreateServiceAccount(name)

	return res.Resources{
		"02-rolebindings/commit-status-tracker-role.yaml":            roles.CreateRole(name, roleRules),
		"02-rolebindings/commit-status-tracker-rolebinding.yaml":     roles.CreateRoleBinding(name, sa, "Role", operatorName),
		"02-rolebindings/commit-status-tracker-service-account.yaml": sa,
		"10-commit-status-tracker/operator.yaml":                     createStatusTrackerDeployment(ns, repoURL, driver),
	}, nil
}

func createStatusTrackerDeployment(ns, repoURL, driver string) *appsv1.Deployment {
	return deployment.Create(commitStatusAppLabel, operatorName, containerImage,
		deployment.ServiceAccount(operatorName),
		deployment.Env(makeEnvironment(repoURL, driver)),
		deployment.Command([]string{operatorName}))
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
		{
			Name:  "STATUS_TRACKER_SECRET",
			Value: CommitStatusTrackerSecret,
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

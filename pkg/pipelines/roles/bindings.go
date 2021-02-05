package roles

import (
	"github.com/redhat-developer/kam/pkg/pipelines/meta"
	corev1 "k8s.io/api/core/v1"
	v1rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	roleTypeMeta               = meta.TypeMeta("Role", "rbac.authorization.k8s.io/v1")
	roleBindingTypeMeta        = meta.TypeMeta("RoleBinding", "rbac.authorization.k8s.io/v1")
	clusterRoleBindingTypeMeta = meta.TypeMeta("ClusterRoleBinding", "rbac.authorization.k8s.io/v1")
	serviceAccountTypeMeta     = meta.TypeMeta("ServiceAccount", "v1")

	clusterRoleTypeMeta = meta.TypeMeta("ClusterRole", "rbac.authorization.k8s.io/v1")
)

// CreateServiceAccount creates and returns a new ServiceAccount in the provided
// namespace.
func CreateServiceAccount(name types.NamespacedName) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta:   serviceAccountTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Secrets:    []corev1.ObjectReference{},
	}
}

// AddSecretToSA grants the provided ServiceAccount access to the named secret.
func AddSecretToSA(sa *corev1.ServiceAccount, secretName string) *corev1.ServiceAccount {
	sa.Secrets = append(sa.Secrets, corev1.ObjectReference{Name: secretName})
	return sa
}

// CreateRoleBinding creates and returns a new RoleBinding given name, sa,
// roleKind, and roleName.
func CreateRoleBinding(name types.NamespacedName, sa *corev1.ServiceAccount, roleKind, roleName string) *v1rbac.RoleBinding {
	return CreateRoleBindingForSubjects(name, roleKind, roleName, []v1rbac.Subject{{Kind: sa.Kind, Name: sa.Name, Namespace: sa.Namespace}})
}

// CreateClusterRoleBinding creates and returns a new ClusterRoleBinding
// associating the service account with a role.
func CreateClusterRoleBinding(name types.NamespacedName, sa *corev1.ServiceAccount, roleKind, roleName string) *v1rbac.ClusterRoleBinding {
	return ClusterRoleBindingForSubjects(name, roleKind, roleName, []v1rbac.Subject{{Kind: sa.Kind, Name: sa.Name, Namespace: sa.Namespace}})
}

// CreateRoleBindingForSubjects creates and returns a RoleBinding with multiple
// subjects.
func CreateRoleBindingForSubjects(name types.NamespacedName, roleKind, roleName string, subjects []v1rbac.Subject) *v1rbac.RoleBinding {
	return &v1rbac.RoleBinding{
		TypeMeta:   roleBindingTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Subjects:   subjects,
		RoleRef: v1rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     roleKind,
			Name:     roleName,
		},
	}
}

// ClusterRoleBindingForSubjects creates and returns a ClusterRoleBinding
// associating the subjects with the specified role.
func ClusterRoleBindingForSubjects(name types.NamespacedName, roleKind, roleName string, subjects []v1rbac.Subject) *v1rbac.ClusterRoleBinding {
	return &v1rbac.ClusterRoleBinding{
		TypeMeta:   clusterRoleBindingTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Subjects:   subjects,
		RoleRef: v1rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     roleKind,
			Name:     roleName,
		},
	}
}

// CreateRole creates and returns a Role given a name and policyRules.
func CreateRole(name types.NamespacedName, policyRules []v1rbac.PolicyRule) *v1rbac.Role {
	return &v1rbac.Role{
		TypeMeta:   roleTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Rules:      policyRules,
	}
}

// CreateClusterRole creates and returns a ClusterRole given a name and policy
// rules.
func CreateClusterRole(name types.NamespacedName, policyRules []v1rbac.PolicyRule) *v1rbac.ClusterRole {
	return &v1rbac.ClusterRole{
		TypeMeta:   clusterRoleTypeMeta,
		ObjectMeta: meta.ObjectMeta(name),
		Rules:      policyRules,
	}
}

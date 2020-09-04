package meta

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TypeMeta creates metav1.TypeMeta
func TypeMeta(kind, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}

// ObjectMeta creates metav1.ObjectMeta
func ObjectMeta(n types.NamespacedName, opts ...objectMetaFunc) metav1.ObjectMeta {
	om := metav1.ObjectMeta{
		Namespace: n.Namespace,
		Name:      n.Name,
	}
	for _, o := range opts {
		o(&om)
	}
	return om
}

// AddLabels is an option func for the ObjectMeta function, which additively
// applies the provided labels to the created ObjectMeta.
func AddLabels(l map[string]string) objectMetaFunc {
	return func(om *metav1.ObjectMeta) {
		if om.Labels == nil {
			om.Labels = map[string]string{}
		}
		for k, v := range l {
			om.Labels[k] = v
		}
	}
}

type objectMetaFunc func(om *metav1.ObjectMeta)

// NamespacedName creates types.NamespacedName
func NamespacedName(ns, name string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}
}

package meta

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TypeMeta creates and returns a new metav1.TypeMeta.
func TypeMeta(kind, apiVersion string) metav1.TypeMeta {
	return metav1.TypeMeta{
		Kind:       kind,
		APIVersion: apiVersion,
	}
}

// ObjectMeta creates and returns a new metav1.ObjectMeta.
func ObjectMeta(n types.NamespacedName, opts ...ObjectMetaOpt) metav1.ObjectMeta {
	om := metav1.ObjectMeta{
		Namespace: n.Namespace,
		Name:      n.Name,
	}
	for _, o := range opts {
		o(&om)
	}
	return om
}

// TODO: Rename this to ObjectMeta

// Meta creates and returns a new ObjectMeta with just the name populated.
func Meta(n string, opts ...ObjectMetaOpt) metav1.ObjectMeta {
	om := metav1.ObjectMeta{
		Name: n,
	}
	for _, o := range opts {
		o(&om)
	}
	return om
}

// AddLabels is an option func for the ObjectMeta function, which additively
// applies the provided labels to the created ObjectMeta.
func AddLabels(l map[string]string) ObjectMetaOpt {
	return func(om *metav1.ObjectMeta) {
		if om.Labels == nil {
			om.Labels = map[string]string{}
		}
		for k, v := range l {
			om.Labels[k] = v
		}
	}
}

// AddAnnotations is an option func for the ObjectMeta function, which additively
// applies the provided labels to the created ObjectMeta.
func AddAnnotations(l map[string]string) ObjectMetaOpt {
	return func(om *metav1.ObjectMeta) {
		if om.Annotations == nil {
			om.Annotations = map[string]string{}
		}
		for k, v := range l {
			om.Annotations[k] = v
		}
	}
}

// ObjectMetaOpt is a function that can change a newly created meta.ObjectMeta
// when it's being created.
type ObjectMetaOpt func(om *metav1.ObjectMeta)

// NamespacedName creates types.NamespacedName
func NamespacedName(ns, name string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: ns,
		Name:      name,
	}
}

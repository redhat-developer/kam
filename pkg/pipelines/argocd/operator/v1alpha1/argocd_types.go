package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArgoCD is the Schema for the argocds API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type ArgoCD struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ArgoCDSpec `json:"spec,omitempty"`
}

// ArgoCDRouteSpec defines the desired state for an OpenShift Route.
type ArgoCDRouteSpec struct {
	// Enabled will toggle the creation of the OpenShift Route.
	Enabled bool `json:"enabled"`
}

// ArgoCDServerSpec defines the options for the ArgoCD Server component.
type ArgoCDServerSpec struct {
	// Route defines the desired state for an OpenShift Route for the Argo CD Server component.
	Route ArgoCDRouteSpec `json:"route,omitempty"`
}

// ArgoCDSpec defines the desired state of ArgoCD
// +k8s:openapi-gen=true
type ArgoCDSpec struct {
	// ResourceExclusions is used to completely ignore entire classes of resource group/kinds.
	ResourceExclusions string `json:"resourceExclusions,omitempty"`

	// Server defines the options for the ArgoCD Server component.
	Server ArgoCDServerSpec `json:"server,omitempty"`
}

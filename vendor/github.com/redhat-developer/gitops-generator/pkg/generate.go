//
// Copyright 2021-2022 Red Hat, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitops

import (
	gitopsv1alpha1 "github.com/redhat-developer/gitops-generator/api/v1alpha1"
	"path/filepath"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/redhat-developer/gitops-generator/pkg/resources"
	"github.com/spf13/afero"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	yaml "github.com/redhat-developer/gitops-generator/pkg/yaml"
)

const (
	kustomizeFileName  = "kustomization.yaml"
	deploymentFileName = "deployment.yaml"
	serviceFileName    = "service.yaml"
	routeFileName      = "route.yaml"
)

var CreatedBy = "application-service"

// Generate takes in a given Component CR and
// spits out a deployment, service, and route file to disk
func Generate(fs afero.Afero, gitOpsFolder string, outputFolder string, component gitopsv1alpha1.Component, commonStorage *corev1.PersistentVolumeClaim) error {
	deployment := generateDeployment(component)

	k := resources.Kustomization{
		APIVersion: "kustomize.config.k8s.io/v1beta1",
		Kind:       "Kustomization",
	}
	k.AddResources(deploymentFileName)
	resources := map[string]interface{}{
		deploymentFileName: deployment,
	}

	// If a targetPort was specified, also generate a service and route
	if component.Spec.TargetPort != 0 {
		service := generateService(component)
		route := generateRoute(component)
		k.AddResources(deploymentFileName, serviceFileName, routeFileName)
		resources[serviceFileName] = service
		resources[routeFileName] = route
	}

	resources[kustomizeFileName] = k

	_, err := yaml.WriteResources(fs, outputFolder, resources)
	if err != nil {
		return err
	}

	// Re-generate the parent kustomize file and return
	return GenerateParentKustomize(fs, gitOpsFolder, commonStorage)
}

// GenerateParentKustomize takes in a folder of components, and outputs a kustomize file to the outputFolder dir
// containing entries for each Component.
// If commonStoragePVC is non-nil, it will also add the common storage pvc yaml file to the parent kustomize. If it's nil, it will not be added
func GenerateParentKustomize(fs afero.Afero, gitOpsFolder string, commonStoragePVC *corev1.PersistentVolumeClaim) error {
	componentsFolder := filepath.Join(gitOpsFolder, "components")
	k := resources.Kustomization{
		Kind:       "Kustomization",
		APIVersion: "kustomize.config.k8s.io/v1beta1",
	}

	resources := map[string]interface{}{}

	fInfo, err := fs.ReadDir(componentsFolder)
	if err != nil {
		return err
	}
	for _, file := range fInfo {
		if file.IsDir() {
			k.AddBases(filepath.Join("components", file.Name(), "base"))
		}
	}

	// if the common storage PVC yaml file was passed in, write to disk and add it to the parent kustomize file
	if commonStoragePVC != nil {
		resources["common-storage-pvc.yaml"] = commonStoragePVC
		k.AddResources("common-storage-pvc.yaml")
	}

	// if the common storage PVC already exist, make sure too add it to the parent kustomize file
	commonStorageExist, err := fs.Exists(filepath.Join(gitOpsFolder, "common-storage-pvc.yaml"))
	if err != nil {
		return err
	}
	if commonStorageExist {
		k.AddResources("common-storage-pvc.yaml")
	}

	resources[kustomizeFileName] = k

	_, err = yaml.WriteResources(fs, gitOpsFolder, resources)
	return err
}

func generateDeployment(component gitopsv1alpha1.Component) *appsv1.Deployment {
	var containerImage string
	if component.Spec.ContainerImage != "" {
		containerImage = component.Spec.ContainerImage
	}
	replicas := getReplicas(component)
	k8sLabels := generateK8sLabels(component)
	matchLabels := getMatchLabel(component)
	deployment := appsv1.Deployment{
		TypeMeta: v1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      component.Name,
			Namespace: component.Namespace,
			Labels:    k8sLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &v1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: matchLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "container-image",
							Image:           containerImage,
							ImagePullPolicy: corev1.PullAlways,
							Env:             component.Spec.Env,
							Resources:       component.Spec.Resources,
						},
					},
				},
			},
		},
	}

	// If a container image source was set in the component *and* a given secret was set for it,
	// Set the secret as an image pull secret, in case the component references a private image component
	if component.Spec.ContainerImage != "" && component.Spec.Secret != "" {
		deployment.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{
				Name: component.Spec.Secret,
			},
		}
	}

	// Set fields that may have been optionally configured by the component CR
	if component.Spec.TargetPort != 0 {
		deployment.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
			{
				ContainerPort: int32(component.Spec.TargetPort),
			},
		}
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{
			InitialDelaySeconds: 10,
			PeriodSeconds:       10,
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(component.Spec.TargetPort),
				},
			},
		}
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{
			InitialDelaySeconds: 10,
			PeriodSeconds:       10,
			ProbeHandler: corev1.ProbeHandler{
				HTTPGet: &corev1.HTTPGetAction{
					Port: intstr.FromInt(component.Spec.TargetPort),
					Path: "/",
				},
			},
		}
	}

	return &deployment
}

func generateService(component gitopsv1alpha1.Component) *corev1.Service {
	k8sLabels := generateK8sLabels(component)
	matchLabels := getMatchLabel(component)
	service := corev1.Service{
		TypeMeta: v1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      component.Name,
			Namespace: component.Namespace,
			Labels:    k8sLabels,
		},
		Spec: corev1.ServiceSpec{
			Selector: matchLabels,
			Ports: []corev1.ServicePort{
				{
					Port:       int32(component.Spec.TargetPort),
					TargetPort: intstr.FromInt(component.Spec.TargetPort),
				},
			},
		},
	}

	return &service
}

func generateRoute(component gitopsv1alpha1.Component) *routev1.Route {
	k8sLabels := generateK8sLabels(component)
	weight := int32(100)
	route := routev1.Route{
		TypeMeta: v1.TypeMeta{
			Kind:       "Route",
			APIVersion: "route.openshift.io/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      component.Name,
			Namespace: component.Namespace,
			Labels:    k8sLabels,
		},
		Spec: routev1.RouteSpec{
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromInt(component.Spec.TargetPort),
			},
			TLS: &routev1.TLSConfig{
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyRedirect,
				Termination:                   routev1.TLSTerminationEdge,
			},
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   component.Name,
				Weight: &weight,
			},
		},
	}

	// If the route field is set in the spec, set it to be the host for the route
	if component.Spec.Route != "" {
		route.Spec.Host = component.Spec.Route
	}

	return &route
}

// getReplicas returns the number of replicas to be created for the component
// If the field is not set, it returns a default value of 1
// ToDo: Handle as part of a defaulting webhook
func getReplicas(component gitopsv1alpha1.Component) int32 {
	if component.Spec.Replicas > 0 {
		return int32(component.Spec.Replicas)
	}
	return 1
}

// generateLabels returns a map containing the following common Kubernetes labels:
// app.kubernetes.io/name: "<component-name>"
// app.kubernetes.io/instance: "<component-cr-name>"
// app.kubernetes.io/part-of: "<application-name>"
// app.kubernetes.io/managed-by: "kustomize"
// app.kubernetes.io/created-by: "application-service"
func generateK8sLabels(component gitopsv1alpha1.Component) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       component.Spec.ComponentName,
		"app.kubernetes.io/instance":   component.Name,
		"app.kubernetes.io/part-of":    component.Spec.Application,
		"app.kubernetes.io/managed-by": "kustomize",
		"app.kubernetes.io/created-by": CreatedBy,
	}
}

// GetMatchLabel returns the label selector that will be used to tie deployments, services, and pods together
// For cleanliness, using just one unique label from the generateK8sLabels function
func getMatchLabel(component gitopsv1alpha1.Component) map[string]string {
	return map[string]string{
		"app.kubernetes.io/instance": component.Name,
	}
}

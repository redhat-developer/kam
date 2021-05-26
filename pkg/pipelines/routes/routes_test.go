package routes

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/redhat-developer/kam/pkg/pipelines/meta"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	testNS   = "testing"
	testName = "test-service"
)

var testSvc = &corev1.Service{
	TypeMeta:   meta.TypeMeta("Service", "v1"),
	ObjectMeta: meta.ObjectMeta(meta.NamespacedName(testNS, testName)),
	Spec: corev1.ServiceSpec{
		Ports: []corev1.ServicePort{
			{
				Name:       "http",
				Protocol:   corev1.ProtocolTCP,
				Port:       8080,
				TargetPort: intstr.FromInt(8080)},
		},
	},
}

func TestNewFromService(t *testing.T) {
	want := map[string]interface{}{
		"apiVersion": "route.openshift.io/v1",
		"kind":       "Route",
		"metadata": map[string]interface{}{
			"creationTimestamp": nil,
			"name":              testName,
			"namespace":         testNS,
		},
		"spec": map[string]interface{}{
			"port": map[string]interface{}{"targetPort": defaultRoutePortName},
			"to": map[string]interface{}{
				"kind":   "Service",
				"name":   testName,
				"weight": float64(100),
			},
			"wildcardPolicy": "None",
		},
	}

	route, err := NewFromService(testSvc)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, route); diff != "" {
		t.Fatalf("NewRoute() failed:\n%s", diff)
	}
}

func TestCreateRoute(t *testing.T) {
	weight := int32(100)
	validRoute := routev1.Route{
		TypeMeta: routeTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      testName,
			Namespace: testNS,
		},
		Spec: routev1.RouteSpec{
			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   testName,
				Weight: &weight,
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString(defaultRoutePortName),
			},
			WildcardPolicy: routev1.WildcardPolicyNone,
		},
	}
	route := createRoute(testSvc)
	if diff := cmp.Diff(validRoute, route); diff != "" {
		t.Fatalf("createRoute() failed:\n%s", diff)
	}
}

func TestCreateRoutePort(t *testing.T) {
	validRoutePort := &routev1.RoutePort{
		TargetPort: intstr.FromString(defaultRoutePortName),
	}
	routePort := createRoutePort(defaultRoutePortName)
	if diff := cmp.Diff(routePort, validRoutePort); diff != "" {
		t.Fatalf("createRoutePort() failed:\n%s", diff)
	}
}

func TestCreatRouteTargetReference(t *testing.T) {
	weight := int32(100)
	validRouteTargetReference := routev1.RouteTargetReference{
		Kind:   "Service",
		Name:   "el-cicd-event-listener",
		Weight: &weight,
	}
	routeTargetReference := creatRouteTargetReference("Service", "el-cicd-event-listener", 100)
	if diff := cmp.Diff(validRouteTargetReference, routeTargetReference); diff != "" {
		t.Fatalf("creatRouteTargetReference() failed:\n%s", diff)
	}
}

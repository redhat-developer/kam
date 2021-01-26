package routes

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/util/intstr"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/redhat-developer/kam/pkg/pipelines/meta"
	corev1 "k8s.io/api/core/v1"
)

var (
	routeTypeMeta = meta.TypeMeta("Route", "route.openshift.io/v1")
)

const defaultRoutePort = 8080

// NewFromService creates and returns an OpenShift route preconfigured for the
// provided Service.
//
// It strips out the Status field from the route as this causes issues when
// being created in a cluster.
func NewFromService(svc *corev1.Service) (interface{}, error) {
	r := createRoute(svc)
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}
	// These are removed because they cause synchronisation issues in ArgoCD.
	delete(result, "status")
	delete(result["spec"].(map[string]interface{}), "host")
	return result, nil
}

func createRoute(svc *corev1.Service) routev1.Route {
	return routev1.Route{
		TypeMeta:   routeTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(svc.Namespace, svc.Name)),
		Spec: routev1.RouteSpec{
			To: creatRouteTargetReference(
				"Service",
				svc.Name,
				100,
			),
			Port:           createRoutePort(svc.Spec.Ports[0].Port),
			WildcardPolicy: routev1.WildcardPolicyNone,
		},
	}
}

func createRoutePort(port int32) *routev1.RoutePort {
	return &routev1.RoutePort{
		TargetPort: intstr.FromInt(int(port)),
	}
}

func creatRouteTargetReference(kind, name string, weight int32) routev1.RouteTargetReference {
	return routev1.RouteTargetReference{
		Kind:   kind,
		Name:   name,
		Weight: &weight,
	}
}

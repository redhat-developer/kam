package eventlisteners

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/util/intstr"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/redhat-developer/kam/pkg/pipelines/meta"
)

// GitOpsWebhookEventListenerRouteName is the OpenShift Route name for GitOps Webhook Listener
const GitOpsWebhookEventListenerRouteName = "gitops-webhook-event-listener-route"

var (
	routeTypeMeta = meta.TypeMeta("Route", "route.openshift.io/v1")
)

const defaultRoutePort = 8080

// GenerateRoute generates an OpenShift route for the EventListener.
//
// It strips out the Status field from the route as this causes issues when
// being created in a cluster.
func GenerateRoute(ns string) (interface{}, error) {
	r := createRoute(ns)
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		return nil, err
	}
	delete(result, "status")
	return result, nil
}

func createRoute(ns string) routev1.Route {
	return routev1.Route{
		TypeMeta:   routeTypeMeta,
		ObjectMeta: meta.ObjectMeta(meta.NamespacedName(ns, GitOpsWebhookEventListenerRouteName)),
		Spec: routev1.RouteSpec{
			To: creatRouteTargetReference(
				"Service",
				"el-cicd-event-listener",
				100,
			),
			Port:           createRoutePort(defaultRoutePort),
			WildcardPolicy: routev1.WildcardPolicyNone,
		},
	}
}

func createRoutePort(port int32) *routev1.RoutePort {
	return &routev1.RoutePort{
		TargetPort: intstr.IntOrString{
			IntVal: port,
		},
	}
}

func creatRouteTargetReference(kind, name string, weight int32) routev1.RouteTargetReference {
	return routev1.RouteTargetReference{
		Kind:   kind,
		Name:   name,
		Weight: &weight,
	}
}

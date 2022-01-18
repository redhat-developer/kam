module github.com/redhat-developer/kam

go 1.16

require (
	github.com/argoproj/argo-cd/v2 v2.1.0
	github.com/code-ready/clicumber v0.0.0-20210201104241-cecb794bdf9a
	github.com/cucumber/godog v0.9.0
	github.com/cucumber/messages-go/v10 v10.0.3
	github.com/evanphx/json-patch v4.12.0+incompatible // indirect
	github.com/google/go-cmp v0.5.6
	github.com/h2non/gock v1.0.9
	github.com/jenkins-x/go-scm v1.10.10
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mkmik/multierror v0.3.0
	github.com/onsi/gomega v1.17.0 // indirect
	github.com/openshift/api v3.9.1-0.20190916204813-cdbe64fb0c91+incompatible
	github.com/openshift/client-go v0.0.0-20200326155132-2a6cd50aedd0
	github.com/openshift/odo v1.2.6
	github.com/operator-framework/api v0.8.0
	github.com/operator-framework/operator-lifecycle-manager v0.18.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.6.0
	github.com/spf13/cobra v1.2.1
	github.com/tektoncd/pipeline v0.28.0
	github.com/tektoncd/triggers v0.16.0
	github.com/zalando/go-keyring v0.1.1
	go.uber.org/goleak v1.1.12 // indirect
	gopkg.in/AlecAivazis/survey.v1 v1.8.8
	k8s.io/api v0.22.2
	k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v11.0.1-0.20190816222228-6d55c1b1f1ca+incompatible
	k8s.io/klog/v2 v2.9.0
	k8s.io/kubectl v0.21.0
	knative.dev/pkg v0.0.0-20210827184538-2bd91f75571c
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/go-openapi/spec => github.com/go-openapi/spec v0.19.8
	github.com/h2non/gock => gopkg.in/h2non/gock.v1 v1.0.14
	k8s.io/api => k8s.io/api v0.21.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0
	k8s.io/apiserver => k8s.io/apiserver v0.21.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.0
	k8s.io/client-go => k8s.io/client-go v0.21.0
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.0
	k8s.io/code-generator => k8s.io/code-generator v0.21.0
	k8s.io/component-base => k8s.io/component-base v0.21.0
	k8s.io/component-helpers => k8s.io/component-helpers v0.21.0
	k8s.io/controller-manager => k8s.io/controller-manager v0.21.0
	k8s.io/cri-api => k8s.io/cri-api v0.21.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.0
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.0
	k8s.io/kubectl => k8s.io/kubectl v0.21.0
	k8s.io/kubelet => k8s.io/kubelet v0.21.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.0
	k8s.io/metrics => k8s.io/metrics v0.21.0
	k8s.io/mount-utils => k8s.io/mount-utils v0.21.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.0
)

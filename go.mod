module github.com/redhat-developer/kam

go 1.16

require (
	github.com/code-ready/clicumber v0.0.0-20210201104241-cecb794bdf9a
	github.com/cucumber/godog v0.9.0
	github.com/cucumber/messages-go/v10 v10.0.3
	github.com/google/go-cmp v0.5.6
	github.com/h2non/gock v1.0.9
	github.com/jenkins-x/go-scm v1.10.10
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mkmik/multierror v0.3.0
	github.com/openshift/api v0.0.0-20210503193030-25175d9d392d
	github.com/openshift/client-go v0.0.0-20210503124028-ac0910aac9fa
	github.com/openshift/odo v1.2.6
	github.com/operator-framework/api v0.8.0
	github.com/operator-framework/operator-lifecycle-manager v0.18.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.6.0
	github.com/spf13/cobra v1.2.1
	github.com/tektoncd/pipeline v0.28.0
	github.com/tektoncd/triggers v0.16.0
	github.com/zalando/go-keyring v0.1.1
	gopkg.in/AlecAivazis/survey.v1 v1.8.8
	k8s.io/api v0.21.4
	k8s.io/apiextensions-apiserver v0.21.4
	k8s.io/apimachinery v0.21.4
	k8s.io/client-go v0.21.4
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.21.0
	knative.dev/pkg v0.0.0-20210827184538-2bd91f75571c
	sigs.k8s.io/yaml v1.2.0
)

replace github.com/h2non/gock => gopkg.in/h2non/gock.v1 v1.0.14

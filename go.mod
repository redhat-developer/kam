module github.com/chetan-rns/gitops-cli

go 1.14

require (
	github.com/Azure/go-autorest/autorest v0.11.4 // indirect
	github.com/Microsoft/go-winio v0.4.15-0.20200113171025-3fe6c5262873 // indirect
	github.com/Netflix/go-expect v0.0.0-20200312175327-da48e75238e2 // indirect
	github.com/bitnami-labs/sealed-secrets v0.12.5
	github.com/docker/docker v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible // indirect
	github.com/docker/go-connections v0.4.1-0.20200120150455-7dc0a2d6ddce // indirect
	github.com/elazarl/goproxy v0.0.0-20190911111923-ecfe977594f1 // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/gobwas/glob v0.2.4-0.20181002190808-e7a84e9525fe // indirect
	github.com/google/go-cmp v0.5.0
	github.com/google/go-querystring v1.0.1-0.20190318165438-c8c88dbee036 // indirect
	github.com/gophercloud/gophercloud v0.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/h2non/gock v1.0.9
	github.com/hinshun/vt10x v0.0.0-20180809195222-d55458df857c // indirect
	github.com/jenkins-x/go-scm v1.5.160
	github.com/kubernetes-sigs/service-catalog v0.2.2
	github.com/kylelemons/godebug v1.1.1-0.20190824192725-fa7b53cdfc91 // indirect
	github.com/mattn/go-isatty v0.0.13-0.20200128103942-cb30d6282491 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mkmik/multierror v0.3.0
	github.com/olekukonko/tablewriter v0.0.0-20180506121414-d4647c9c7a84 // indirect
	github.com/onsi/ginkgo v1.14.0 // indirect
	github.com/opencontainers/image-spec v1.0.2-0.20200206005212-79b036d80240 // indirect
	github.com/openshift/api v3.9.1-0.20190924102528-32369d4db2ad+incompatible
	github.com/openshift/client-go v0.0.0-20200116152001-92a2713fa240
	github.com/openshift/library-go v0.0.0-20200407165825-2e79bd232e72 // indirect
	github.com/openshift/odo v1.2.6
	github.com/operator-framework/operator-lifecycle-manager v0.0.0-20200422144016-a6acf50218ed // indirect
	github.com/pkg/errors v0.9.1
	github.com/posener/complete v1.1.2
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/tektoncd/pipeline v0.15.2
	github.com/tektoncd/triggers v0.7.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zalando/go-keyring v0.1.0 // indirect
	gopkg.in/AlecAivazis/survey.v1 v1.8.0
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.17.2
	knative.dev/pkg v0.0.0-20200702222342-ea4d6e985ba0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/apcera/gssapi => github.com/openshift/gssapi v0.0.0-20161010215902-5fb4217df13b
	github.com/containers/image => github.com/openshift/containers-image v0.0.0-20190130162819-76de87591e9d
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200309214505-aa6a9891b09c+incompatible
	github.com/h2non/gock => gopkg.in/h2non/gock.v1 v1.0.14
	github.com/openshift/api => github.com/openshift/api v0.0.0-20200205133042-34f0ec8dab87
	k8s.io/api => k8s.io/api v0.17.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.1
	k8s.io/apimachinery => github.com/openshift/kubernetes-apimachinery v0.0.0-20191211181342-5a804e65bdc1
	k8s.io/apiserver => k8s.io/apiserver v0.17.1
	k8s.io/cli-runtime => github.com/openshift/kubernetes-cli-runtime v0.0.0-20200114162348-c8810ef308ee
	k8s.io/client-go => github.com/openshift/kubernetes-client-go v0.0.0-20191211181558-5dcabadb2b45
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.1
	k8s.io/code-generator => k8s.io/code-generator v0.17.1
	k8s.io/component-base => k8s.io/component-base v0.17.1
	k8s.io/cri-api => k8s.io/cri-api v0.17.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.1
	k8s.io/kubectl => github.com/openshift/kubernetes-kubectl v0.0.0-20200211153013-50adac736181
	k8s.io/kubelet => k8s.io/kubelet v0.17.1
	k8s.io/kubernetes => github.com/openshift/kubernetes v1.17.0-alpha.0.0.20191216151305-079984b0a154
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.1
	k8s.io/metrics => k8s.io/metrics v0.17.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.1

)

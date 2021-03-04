package cmd

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/google/go-cmp/cmp"
	v1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	operatorsfake "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/fake"
	"github.com/redhat-developer/kam/pkg/cmd/utility"
	"github.com/redhat-developer/kam/pkg/pipelines"
	"github.com/redhat-developer/kam/pkg/pipelines/argocd"
	"github.com/redhat-developer/kam/pkg/pipelines/secrets"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
)

type mockSpinner struct {
	writer io.Writer
	start  bool
	end    bool
}

const (
	gitOpsURL  = "https://github.com/org/gitops"
	serviceURL = "https://github.com/org/app"
)

const (
	customSealedSecretsNS         = "sealed-secrets"
	customSealedSecretsController = "custom-sealed-secrets-controller"
)

func TestValidatePrefix(t *testing.T) {
	completeTests := []struct {
		name        string
		prefix      string
		wantPrefix  string
		serviceRepo string
		gitRepo     string
	}{
		{"no prefix", "", "", "https://github.com/gaganhegde/test-repo.git", "https://github.com/gaganhegde/taxi.git"},
		{"prefix with hyphen", "test-", "test-", "https://github.com/gaganhegde/test-repo.git", "https://github.com/gaganhegde/taxi.git"},
		{"prefix without hyphen", "test", "test-", "https://github.com/gaganhegde/test-repo.git", "https://github.com/gaganhegde/taxi.git"},
	}

	for _, tt := range completeTests {
		o := BootstrapParameters{
			BootstrapOptions: &pipelines.BootstrapOptions{
				Prefix: tt.prefix, GitOpsRepoURL: tt.gitRepo,
				ServiceRepoURL: tt.serviceRepo, ImageRepo: ""},
		}

		err := o.Validate()

		if err != nil {
			t.Errorf("Validate() %#v failed: ", err)
		}

		if o.Prefix != tt.wantPrefix {
			t.Errorf("Validate() %#v prefix: got %s, want %s", tt.name, o.Prefix, tt.wantPrefix)
		}
	}
}

func TestAddSuffixWithBootstrap(t *testing.T) {
	tt := []struct {
		name           string
		gitOpsURL      string
		appURL         string
		validGitOpsURL string
		validAppURL    string
	}{
		{"suffix already exists", gitOpsURL + ".git", serviceURL + ".git", gitOpsURL + ".git", serviceURL + ".git"},
		{"misssing suffix", gitOpsURL, serviceURL, gitOpsURL + ".git", serviceURL + ".git"},
	}

	for _, test := range tt {
		t.Run(test.name, func(rt *testing.T) {
			o := &BootstrapParameters{
				BootstrapOptions: &pipelines.BootstrapOptions{
					GitOpsRepoURL:  test.gitOpsURL,
					ServiceRepoURL: test.appURL},
			}

			addGitURLSuffixIfNecessary(o)

			if o.GitOpsRepoURL != test.validGitOpsURL {
				rt.Fatalf("URL mismatch: got %s, want %s", o.GitOpsRepoURL, test.validAppURL)
			}
			if o.ServiceRepoURL != test.validAppURL {
				rt.Fatalf("URL mismatch: got %s, want %s", o.GitOpsRepoURL, test.validAppURL)
			}
		})
	}
}

func TestValidateCommitStatusTracker(t *testing.T) {
	completeTests := []struct {
		name                string
		gitRepo             string
		commitStatusTracker bool
		gitAccessToken      string
		wantErr             string
	}{
		{"statusTracker true/ GitAccessToken absent", "username1/testRepo1", true, "", "--git-host-access-token is required if commit-status-tracker is enabled"},
		{"statusTracker true/ GitAccessToken present", "username2/testRepo2", true, "abc123", ""},
		{"statusTracker false/ GitAccessToken present", "username3/testRepo3", false, "abc123", ""},
		{"statusTracker false/ GitAccessToken present", "username3/testRepo3", false, "abc123", ""},
	}

	for _, tt := range completeTests {
		o := BootstrapParameters{
			BootstrapOptions: &pipelines.BootstrapOptions{
				GitOpsRepoURL:       tt.gitRepo,
				CommitStatusTracker: tt.commitStatusTracker,
				GitHostAccessToken:  tt.gitAccessToken,
			},
		}

		got := o.Validate()
		gotErr := ""
		if got != nil {
			gotErr = got.Error()
		}
		if diff := cmp.Diff(tt.wantErr, gotErr); diff != "" {
			t.Fatalf("Validate() for case %s didn't match: %s\n", tt.name, diff)
		}
	}

}

func TestValidateBootstrapParameter(t *testing.T) {
	optionTests := []struct {
		name    string
		gitRepo string
		driver  string
		errMsg  string
	}{
		{"invalid repo", "test", "", "repo must be org/repo"},
		{"valid repo", "test/repo", "", ""},
		{"invalid driver", "test/repo", "unknown", "invalid"},
		{"valid driver gitlab", "test/repo", "gitlab", ""},
	}

	for _, tt := range optionTests {
		o := BootstrapParameters{
			BootstrapOptions: &pipelines.BootstrapOptions{
				GitOpsRepoURL:     tt.gitRepo,
				PrivateRepoDriver: tt.driver,
				Prefix:            "test",
			},
		}
		err := o.Validate()

		if err != nil && tt.errMsg == "" {
			t.Errorf("Validate() %#v got an unexpected error: %s", tt.name, err)
			continue
		}

		if !matchError(t, tt.errMsg, err) {
			t.Errorf("Validate() %#v failed to match error: got %s, want %s", tt.name, err, tt.errMsg)
		}
	}
}

func TestValidatePairFlags(t *testing.T) {
	optionTests := []struct {
		name          string
		token         string
		statustracker bool
		keyring       bool
		errMsg        string
	}{
		{"--save-token-keyring set and --git-host-access-token missing", "", false, true, "--git-host-access-token is required if --save-token-keyring is enabled"},
		{"--commit-status-tracker set and --git-host-access-token missing", "", true, false, "--git-host-access-token is required if commit-status-tracker is enabled"},
		{"--commit-status-tracker/--save-token-keyring set and --git-host-access-token present", "abc123", true, true, ""},
		{"--commit-status-tracker/--save-token-keyring not-set and --git-host-access-token absent", "", false, false, ""},
	}

	for _, tt := range optionTests {
		o := BootstrapParameters{
			BootstrapOptions: &pipelines.BootstrapOptions{
				GitOpsRepoURL:       gitOpsURL,
				ServiceRepoURL:      serviceURL,
				ImageRepo:           "io/test/repo",
				GitHostAccessToken:  tt.token,
				CommitStatusTracker: tt.statustracker,
				SaveTokenKeyRing:    tt.keyring,
			},
		}
		err := o.Validate()

		if err != nil && tt.errMsg == "" {
			t.Errorf("Validate() %#v got an unexpected error: %s", tt.name, err)
			continue
		}

		if !matchError(t, tt.errMsg, err) {
			t.Errorf("Validate() %#v failed to match error: got %s, want %s", tt.name, err, tt.errMsg)
		}
	}
}

func TestValidateMandatoryFlags(t *testing.T) {
	optionTests := []struct {
		name                string
		gitRepo             string
		serviceRepo         string
		imagerepo           string
		commitStatusTracker bool
		gitToken            string
		errMsg              string
	}{
		{"missing gitops-repo-url", "", "https://github.com/example/repo.git", "registry/username/repo", false, "", `required flag(s) "gitops-repo-url" not set`},
		{"missing service-repo-url", "https://github.com/example/repo.git", "", "registry/username/repo", false, "", `required flag(s) "service-repo-url" not set`},
	}

	for _, tt := range optionTests {
		o := BootstrapParameters{
			BootstrapOptions: &pipelines.BootstrapOptions{
				GitOpsRepoURL:  tt.gitRepo,
				ServiceRepoURL: tt.serviceRepo,
				ImageRepo:      tt.imagerepo,
			},
		}
		err := nonInteractiveMode(&o, &utility.Client{})
		if tt.errMsg != err.Error() {
			t.Fatalf("nonInteractiveMode() %#v failed to match error: got %s, want %s", tt.name, err, tt.errMsg)
		}
	}
}
func TestCheckSpinner(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		endStatus bool
		wantMsg   string
	}{
		{
			"Resource not found error",
			errors.NewNotFound(schema.GroupResource{}, "abcd"),
			false,
			"\nChecking if abcd is installed [Please install abcd]",
		},
		{
			"Random cluster error",
			fmt.Errorf("Sample cluster error"),
			false,
			"\nChecking if abcd is installed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buff := &bytes.Buffer{}

			fakeSpinner := &mockSpinner{writer: buff}
			fakeSpinner.Start("Checking if abcd is installed", false)
			warnIfNotFound(fakeSpinner, "Please install abcd", test.err)

			if fakeSpinner.end != test.endStatus {
				t.Errorf("Spinner status mismatch: got %v, want %v", fakeSpinner.end, test.endStatus)
			}
			assertMessage(t, buff.String(), test.wantMsg)
		})
	}
}

func TestDependenciesWithNothingInstalled(t *testing.T) {
	fakeClient := newFakeClient(nil, nil)

	wantMsg := `
Checking if Sealed Secrets is installed with the default configuration [The Sealed Secrets Controller was not detected]
Checking if ArgoCD is installed with the default configuration [Please install OpenShift GitOps Operator from OperatorHub]
Checking if OpenShift Pipelines Operator is installed with the default configuration [Please install OpenShift GitOps Operator from OperatorHub]`

	buff := &bytes.Buffer{}
	fakeSpinner := &mockSpinner{writer: buff}
	err := checkBootstrapDependencies(
		&BootstrapParameters{BootstrapOptions: &pipelines.BootstrapOptions{SealedSecretsService: types.NamespacedName{Namespace: secrets.SealedSecretsNS, Name: secrets.SealedSecretsController}}},
		fakeClient, fakeSpinner)
	wantErr := fmt.Sprintf("failed to satisfy the required dependencies: %s, %s", gitopsOperatorName, pipelinesOperatorName)

	assertError(t, err, wantErr)
	assertMessage(t, buff.String(), wantMsg)
}

func TestDependenciesWithAllInstalled(t *testing.T) {
	fakeClient := newFakeClient([]runtime.Object{defaultSealedSecretsService(), pipelinesOperator()}, []runtime.Object{argoCDCSV()})

	wantMsg := `
Checking if Sealed Secrets is installed with the default configuration
Checking if ArgoCD is installed with the default configuration
Checking if OpenShift Pipelines Operator is installed with the default configuration`

	buff := &bytes.Buffer{}
	fakeSpinner := &mockSpinner{writer: buff}
	wizardParams := &BootstrapParameters{BootstrapOptions: &pipelines.BootstrapOptions{}}
	err := checkBootstrapDependencies(wizardParams, fakeClient, fakeSpinner)

	assertError(t, err, "")
	if wizardParams.SealedSecretsService.Name != secrets.SealedSecretsController && wizardParams.SealedSecretsService.Namespace != secrets.SealedSecretsNS {
		t.Fatalf("Expected sealed secrets to be set")
	}
	assertMessage(t, buff.String(), wantMsg)
}

func TestDependenciesWithAllInstalledDifferentSealedSecretsService(t *testing.T) {
	// use a custom sealed secret service
	fakeClient := newFakeClient([]runtime.Object{customSealedSecretsService(), pipelinesOperator()}, []runtime.Object{argoCDCSV()})

	// expect negative Sealed Secrets check
	wantMsg := `
Checking if Sealed Secrets is installed with the default configuration [The Sealed Secrets Controller was not detected]
Checking if ArgoCD is installed with the default configuration
Checking if OpenShift Pipelines Operator is installed with the default configuration`

	buff := &bytes.Buffer{}
	fakeSpinner := &mockSpinner{writer: buff}
	// NO parameter for the custom sealed secret service defined
	wizardParams := &BootstrapParameters{BootstrapOptions: &pipelines.BootstrapOptions{}}
	err := checkBootstrapDependencies(wizardParams, fakeClient, fakeSpinner)

	assertError(t, err, "")
	assertMessage(t, buff.String(), wantMsg)
}

func TestDependenciesWithAllInstalledCustomSealedSecretsService(t *testing.T) {
	// use a custom sealed secret service
	fakeClient := newFakeClient([]runtime.Object{customSealedSecretsService(), pipelinesOperator()}, []runtime.Object{argoCDCSV()})

	// expect checking and finding the custom Sealed Secrets config
	wantMsg := `
Checking if Sealed Secrets is installed with custom configuration
Checking if ArgoCD is installed with the default configuration
Checking if OpenShift Pipelines Operator is installed with the default configuration`

	buff := &bytes.Buffer{}
	fakeSpinner := &mockSpinner{writer: buff}
	// set parameter for the custom sealed secret service
	wizardParams := &BootstrapParameters{BootstrapOptions: &pipelines.BootstrapOptions{SealedSecretsService: types.NamespacedName{Namespace: customSealedSecretsNS, Name: customSealedSecretsController}}}
	err := checkBootstrapDependencies(wizardParams, fakeClient, fakeSpinner)

	assertError(t, err, "")
	if wizardParams.SealedSecretsService.Name != customSealedSecretsController && wizardParams.SealedSecretsService.Namespace != customSealedSecretsNS {
		t.Fatalf("Expected sealed secrets to be set")
	}
	assertMessage(t, buff.String(), wantMsg)
}

func TestDependenciesWithAllInstalledCustomSealedSecretsServiceButDefaultIsInstalled(t *testing.T) {
	// use a DEFAULT sealed secret service
	fakeClient := newFakeClient([]runtime.Object{defaultSealedSecretsService(), pipelinesOperator()}, []runtime.Object{argoCDCSV()})

	// expect checking custom Sealed Secrets, but unsuccessful
	wantMsg := `
Checking if Sealed Secrets is installed with custom configuration [Provided Sealed Secrets namespace/name are not valid. Please verify]
Checking if ArgoCD is installed with the default configuration
Checking if OpenShift Pipelines Operator is installed with the default configuration`

	buff := &bytes.Buffer{}
	fakeSpinner := &mockSpinner{writer: buff}
	// set parameter for the custom sealed secret service
	wizardParams := &BootstrapParameters{BootstrapOptions: &pipelines.BootstrapOptions{SealedSecretsService: types.NamespacedName{Namespace: customSealedSecretsNS, Name: customSealedSecretsController}}}
	err := checkBootstrapDependencies(wizardParams, fakeClient, fakeSpinner)

	assertError(t, err, "")
	assertMessage(t, buff.String(), wantMsg)
	// not the default Sealed Secrets are expected
	if wizardParams.SealedSecretsService.Name == secrets.SealedSecretsController || wizardParams.SealedSecretsService.Namespace == secrets.SealedSecretsNS {
		t.Fatalf("Expected sealed secrets to be set")
	}
}

func TestDependenciesWithAllInstalledCustomSealedSecretsServiceButNotMatched(t *testing.T) {
	// use an OTHER sealed secret service
	fakeClient := newFakeClient([]runtime.Object{customSealedSecretsServiceOther(), pipelinesOperator()}, []runtime.Object{argoCDCSV()})

	wantMsg := `
Checking if Sealed Secrets is installed with custom configuration [Provided Sealed Secrets namespace/name are not valid. Please verify]
Checking if ArgoCD is installed with the default configuration
Checking if OpenShift Pipelines Operator is installed with the default configuration`

	buff := &bytes.Buffer{}
	fakeSpinner := &mockSpinner{writer: buff}
	// set parameter for the custom sealed secret service
	wizardParams := &BootstrapParameters{BootstrapOptions: &pipelines.BootstrapOptions{SealedSecretsService: types.NamespacedName{Namespace: customSealedSecretsNS, Name: customSealedSecretsController}}}
	err := checkBootstrapDependencies(wizardParams, fakeClient, fakeSpinner)

	assertError(t, err, "")
	assertMessage(t, buff.String(), wantMsg)
}

func TestDependenciesWithNoArgoCD(t *testing.T) {
	fakeClient := newFakeClient([]runtime.Object{defaultSealedSecretsService(), pipelinesOperator()}, nil)

	wantMsg := `
Checking if Sealed Secrets is installed with the default configuration
Checking if ArgoCD is installed with the default configuration [Please install OpenShift GitOps Operator from OperatorHub]
Checking if OpenShift Pipelines Operator is installed with the default configuration`

	buff := &bytes.Buffer{}
	fakeSpinner := &mockSpinner{writer: buff}
	wizardParams := &BootstrapParameters{
		BootstrapOptions: &pipelines.BootstrapOptions{},
	}
	err := checkBootstrapDependencies(wizardParams, fakeClient, fakeSpinner)
	wantErr := fmt.Sprintf("failed to satisfy the required dependencies: %s", gitopsOperatorName)

	assertError(t, err, wantErr)
	assertMessage(t, buff.String(), wantMsg)
}

func TestDependenciesWithNoPipelines(t *testing.T) {
	fakeClient := newFakeClient([]runtime.Object{defaultSealedSecretsService()}, []runtime.Object{argoCDCSV()})

	wantMsg := `
Checking if Sealed Secrets is installed with the default configuration
Checking if ArgoCD is installed with the default configuration
Checking if OpenShift Pipelines Operator is installed with the default configuration [Please install OpenShift GitOps Operator from OperatorHub]`

	buff := &bytes.Buffer{}
	fakeSpinner := &mockSpinner{writer: buff}
	wizardParams := &BootstrapParameters{BootstrapOptions: &pipelines.BootstrapOptions{}}
	err := checkBootstrapDependencies(wizardParams, fakeClient, fakeSpinner)
	wantErr := fmt.Sprintf("failed to satisfy the required dependencies: %s", pipelinesOperatorName)

	assertError(t, err, wantErr)
	assertMessage(t, buff.String(), wantMsg)
}

func assertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		if msg != "" {
			t.Fatalf("error mismatch: got %v, want %v", err, msg)
		}
		return
	}
	if err.Error() != msg {
		t.Fatalf("error mismatch: got %s, want %s", err.Error(), msg)
	}
}

func assertMessage(t *testing.T, got, want string) {
	t.Helper()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("message comparison failed:\n%s", diff)
	}
}

func defaultSealedSecretsService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secrets.SealedSecretsController,
			Namespace: secrets.SealedSecretsNS,
		},
	}
}

func customSealedSecretsService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      customSealedSecretsController,
			Namespace: customSealedSecretsNS,
		},
	}
}

func customSealedSecretsServiceOther() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dummy-controller-do-not-use-value",
			Namespace: "dummy-ns-do-not-use-value",
		},
	}
}

func pipelinesOperator() *appv1.Deployment {
	return &appv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openshift-pipelines-operator",
			Namespace: "openshift-operators",
		},
	}
}
func matchError(t *testing.T, s string, e error) bool {
	t.Helper()
	if s == "" && e == nil {
		return true
	}
	if s != "" && e == nil {
		return false
	}
	match, err := regexp.MatchString(s, e.Error())
	if err != nil {
		t.Fatal(err)
	}
	return match
}

func (m *mockSpinner) Start(status string, debug bool) {
	m.start = true
	fmt.Fprintf(m.writer, "\n%s", status)
}

func (m *mockSpinner) End(status bool) {
	m.end = status
}

func (m *mockSpinner) WarningStatus(status string) {
	fmt.Fprintf(m.writer, " [%s]", status)
}

func newFakeClient(k8sObjs []runtime.Object, clientObjs []runtime.Object) *utility.Client {
	return &utility.Client{
		KubeClient:     fake.NewSimpleClientset(k8sObjs...),
		OperatorClient: operatorsfake.NewSimpleClientset(clientObjs...).OperatorsV1alpha1(),
	}
}

func argoCDCSV() *v1alpha1.ClusterServiceVersion {
	return &v1alpha1.ClusterServiceVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argocd",
			Namespace: argocd.ArgoCDNamespace,
		},
		Spec: v1alpha1.ClusterServiceVersionSpec{
			CustomResourceDefinitions: v1alpha1.CustomResourceDefinitions{
				Owned: []v1alpha1.CRDDescription{
					{Name: "argocds.argoproj.io", Kind: "ArgoCD"},
					{Name: "fake.crd", Kind: "ArgoCD"},
				},
			},
		},
	}
}

func TestMissingFlags(t *testing.T) {
	tests := []struct {
		desc  string
		flags map[string]string
		err   error
	}{
		{
			"Required flags are present",
			map[string]string{"gitops-repo-url": "value-1", "service-repo-url": "value-2"},
			nil,
		},
		{
			"A required flag is absent",
			map[string]string{"gitops-repo-url": "value-1", "service-repo-url": ""},
			missingFlagErr([]string{`"service-repo-url"`}),
		},
		{
			"Multiple required flags are absent",
			map[string]string{"gitops-repo-url": "", "service-repo-url": ""},
			missingFlagErr([]string{`"service-repo-url"`, `"gitops-repo-url"`}),
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			gotErr := checkMandatoryFlags(test.flags)
			if gotErr != nil && test.err != nil {
				if gotErr.Error() != test.err.Error() {
					t.Fatalf("error mismatch: got %v, want %v", gotErr, test.err)
				}
			} else if gotErr != test.err {
				t.Fatalf("error mismatch: got %v, want %v", gotErr, test.err)
			}
		})
	}
}

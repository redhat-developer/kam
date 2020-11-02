package cmd

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/google/go-cmp/cmp"
	v1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	operatorsfake "github.com/operator-framework/operator-lifecycle-manager/pkg/api/client/clientset/versioned/fake"
	"github.com/redhat-developer/kam/pkg/cmd/utility"
	"github.com/redhat-developer/kam/pkg/pipelines"
	"github.com/redhat-developer/kam/pkg/pipelines/secrets"
	"github.com/zalando/go-keyring"
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
	gitOpsURL := "https://github.com/org/gitops"
	appURL := "https://github.com/org/app"
	tt := []struct {
		name           string
		gitOpsURL      string
		appURL         string
		validGitOpsURL string
		validAppURL    string
	}{
		{"suffix already exists", gitOpsURL + ".git", appURL + ".git", gitOpsURL + ".git", appURL + ".git"},
		{"misssing suffix", gitOpsURL, appURL, gitOpsURL + ".git", appURL + ".git"},
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
		{"missing image-repo", "https://github.com/example/repo.git", "https://github.com/example/repo.git", "", false, "", `required flag(s) "image-repo" not set`},
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

func TestKeyRingSet(t *testing.T) {
	keyring.MockInit()
	optionTests := []struct {
		name          string
		gitRepo       string
		serviceRepo   string
		imagerepo     string
		gitToken      string
		expectedToken string
	}{
		{"exclude gitToken to execute as exepected", "https://github.com/example/gitRepo.git", "https://github.com/example/serviceRepo.git", "registry/username/repo", "", ""},
		{"add git accessToken", "https://github.com/example/gitRepo.git", "https://github.com/example/service.git", "registry/username/repo", "abc123", "abc123"},
		{"overwrite gitops repo access token", "https://github.com/example/gitRepo.git", "https://github.com/example/service.git", "registry/username/repo", "xyz123", "xyz123"},
	}

	for _, tt := range optionTests {
		o := BootstrapParameters{
			BootstrapOptions: &pipelines.BootstrapOptions{
				GitOpsRepoURL:      tt.gitRepo,
				ServiceRepoURL:     tt.serviceRepo,
				ImageRepo:          tt.imagerepo,
				GitHostAccessToken: tt.gitToken,
			},
		}
		err := nonInteractiveMode(&o, &utility.Client{})
		if err != nil {
			t.Fatalf("Non Interactive mode failed with error: %v", err)
		}
		gitopsToken, err := keyring.Get("kam", tt.gitRepo)
		if err != nil {
			t.Fatal(err)
		}
		serviceToken, err := keyring.Get("kam", tt.serviceRepo)
		if err != nil {
			t.Fatal(err)
		}
		if tt.expectedToken != gitopsToken && tt.expectedToken != serviceToken {
			t.Fatalf("TestKeyRingSet() Failed since expected token %v did not match %v", tt.expectedToken, gitopsToken)
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
Checking if Sealed Secrets is installed with the default configuration [Please install Sealed Secrets Operator from OperatorHub]
Checking if ArgoCD Operator is installed with the default configuration [Please install ArgoCD Operator from OperatorHub]
Checking if OpenShift Pipelines Operator is installed with the default configuration [Please install OpenShift Pipelines Operator from OperatorHub]`

	buff := &bytes.Buffer{}
	fakeSpinner := &mockSpinner{writer: buff}
	err := checkBootstrapDependencies(
		&BootstrapParameters{BootstrapOptions: &pipelines.BootstrapOptions{}},
		fakeClient, fakeSpinner)
	wantErr := fmt.Sprintf("failed to satisfy the required dependencies: %s, %s", argoCdOperatorName, pipelinesOperatorName)

	assertError(t, err, wantErr)
	assertMessage(t, buff.String(), wantMsg)
}

func TestDependenciesWithAllInstalled(t *testing.T) {
	fakeClient := newFakeClient([]runtime.Object{sealedSecretsService(), pipelinesOperator()}, []runtime.Object{argoCDCSV()})

	wantMsg := `
Checking if Sealed Secrets is installed with the default configuration
Checking if ArgoCD Operator is installed with the default configuration
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

func TestDependenciesWithNoArgoCD(t *testing.T) {
	fakeClient := newFakeClient([]runtime.Object{sealedSecretsService(), pipelinesOperator()}, nil)

	wantMsg := `
Checking if Sealed Secrets is installed with the default configuration
Checking if ArgoCD Operator is installed with the default configuration [Please install ArgoCD Operator from OperatorHub]
Checking if OpenShift Pipelines Operator is installed with the default configuration`

	buff := &bytes.Buffer{}
	fakeSpinner := &mockSpinner{writer: buff}
	wizardParams := &BootstrapParameters{
		BootstrapOptions: &pipelines.BootstrapOptions{},
	}
	err := checkBootstrapDependencies(wizardParams, fakeClient, fakeSpinner)
	wantErr := fmt.Sprintf("failed to satisfy the required dependencies: %s", argoCdOperatorName)

	assertError(t, err, wantErr)
	assertMessage(t, buff.String(), wantMsg)
}

func TestDependenciesWithNoPipelines(t *testing.T) {
	fakeClient := newFakeClient([]runtime.Object{sealedSecretsService()}, []runtime.Object{argoCDCSV()})

	wantMsg := `
Checking if Sealed Secrets is installed with the default configuration
Checking if ArgoCD Operator is installed with the default configuration
Checking if OpenShift Pipelines Operator is installed with the default configuration [Please install OpenShift Pipelines Operator from OperatorHub]`

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

func sealedSecretsService() *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secrets.SealedSecretsController,
			Namespace: secrets.SealedSecretsNS,
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
			Namespace: "argocd",
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
			map[string]string{"gitops-repo-url": "value-1", "service-repo-url": "value-2", "image-repo": "value-3"},
			nil,
		},
		{
			"A required flag is absent",
			map[string]string{"gitops-repo-url": "value-1", "service-repo-url": "value-2", "image-repo": ""},
			missingFlagErr([]string{`"image-repo"`}),
		},
		{
			"Multiple required flags are absent",
			map[string]string{"gitops-repo-url": "value-1", "service-repo-url": "", "image-repo": ""},
			missingFlagErr([]string{`"service-repo-url"`, `"image-repo"`}),
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

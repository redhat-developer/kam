package utility

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAddGitSuffix(t *testing.T) {
	addSuffixTests := []struct {
		name string
		url  string
		want string
	}{
		{"missing git suffix", "https://github.com/test/org", "https://github.com/test/org.git"},
		{"suffix for empty string", "", ""},
		{"suffix already present", "https://github.com/test/org.git", "https://github.com/test/org.git"},
		{"suffix with a different case", "https://github.com/test/org.GIT", "https://github.com/test/org.GIT"},
	}

	for _, tt := range addSuffixTests {
		t.Run(tt.name, func(rt *testing.T) {
			got := AddGitSuffixIfNecessary(tt.url)
			if tt.want != got {
				rt.Fatalf("URL mismatch: got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestRemoveEmptyStrings(t *testing.T) {
	stringsTests := []struct {
		name   string
		source []string
		want   []string
	}{
		{"no strings", []string{}, []string{}},
		{"no empty strings", []string{"test1", "test2"}, []string{"test1", "test2"}},
		{"mixed strings", []string{"", "test2", ""}, []string{"test2"}},
	}

	for _, tt := range stringsTests {
		t.Run(tt.name, func(rt *testing.T) {
			got := RemoveEmptyStrings(tt.source)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				rt.Fatalf("string removal failed:\n%s", diff)
			}
		})
	}
}

func TestMaybeCompletePrefix(t *testing.T) {
	stringsTests := []struct {
		name   string
		prefix string
		want   string
	}{
		{"with dash on end", "testing-", "testing-"},
		{"with no dash on end", "testing", "testing-"},
	}

	for _, tt := range stringsTests {
		t.Run(tt.name, func(rt *testing.T) {
			got := MaybeCompletePrefix(tt.prefix)
			if tt.want != got {
				rt.Fatalf("prefixing failed, got %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCheckIfSealedSecretsExists(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset(&v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sealed-secrets-controller",
			Namespace: "kube-system",
		},
	})

	fakeClient := Client{KubeClient: fakeClientSet}
	err := fakeClient.CheckIfSealedSecretsExists(types.NamespacedName{Namespace: "kube-system", Name: "sealed-secrets-controller"})
	if err != nil {
		t.Fatalf("CheckIfSealedSecretsExists failed: got %v,want %v", err, nil)
	}
	err = fakeClient.CheckIfSealedSecretsExists(types.NamespacedName{Namespace: "unknown", Name: "unknown"})
	wantErr := `services "unknown" not found`
	if err == nil {
		t.Fatalf("CheckIfSealedSecretsExists failed: got %v,want %v", nil, wantErr)
	}
}

func TestCheckIfArgoCDExists(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argocd-operator",
			Namespace: "argocd",
		},
	}, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argocd-server",
			Namespace: "argocd",
		},
	})

	fakeClient := Client{KubeClient: fakeClientSet}

	err := fakeClient.CheckIfArgoCDExists("argocd")
	if err != nil {
		t.Fatalf("CheckIfArgoCDExists failed: got %v,want %v", err, nil)
	}
	err = fakeClient.CheckIfArgoCDExists("unknown")
	wantErr := `deployments "unknown" not found`
	if err == nil {
		t.Fatalf("CheckIfArgoCDExists failed: got %v, want %v", nil, wantErr)
	}
}

func TestCheckIfPipelinesExists(t *testing.T) {
	fakeClientSet := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "openshift-pipelines-operator",
			Namespace: "openshift-operators",
		},
	})

	fakeClient := Client{KubeClient: fakeClientSet}

	err := fakeClient.CheckIfPipelinesExists("openshift-operators")
	if err != nil {
		t.Fatalf("CheckIfPipelinesExists failed: got %v,want %v", err, nil)
	}
	err = fakeClient.CheckIfPipelinesExists("unknown")
	wantErr := `deployments "unknown" not found`
	if err == nil {
		t.Fatalf("CheckIfPipelinesExists failed: got %v,want %v", nil, wantErr)
	}
}

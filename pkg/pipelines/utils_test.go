package pipelines

import (
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/driver/fake"
)

func TestBootstrapRepository_with_personal_account(t *testing.T) {
	token := "this-is-a-test-token"
	fakeData := stubOutGitClientFactory(t, token)
	fakeData.CurrentUser = scm.User{Login: "testing"}

	err := BootstrapRepository(&BootstrapOptions{
		GitOpsRepoURL:      "https://example.com/testing/test-repo.git",
		GitHostAccessToken: token,
	})
	assertNoError(t, err)

	assertRepositoryCreated(t, fakeData, "", "test-repo")
}

func TestBootstrapRepository_with_org(t *testing.T) {
	token := "this-is-a-test-token"
	fakeData := stubOutGitClientFactory(t, token)
	fakeData.CurrentUser = scm.User{Login: "test-user"}

	err := BootstrapRepository(&BootstrapOptions{
		GitOpsRepoURL:      "https://example.com/testing/test-repo.git",
		GitHostAccessToken: token,
	})
	assertNoError(t, err)
	assertRepositoryCreated(t, fakeData, "testing", "test-repo")
}

func TestBootstrapRepository_with_no_access_token(t *testing.T) {
	token := "this-is-a-test-token"
	fakeData := stubOutGitClientFactory(t, token)
	fakeData.CurrentUser = scm.User{Login: "test-user"}

	err := BootstrapRepository(&BootstrapOptions{
		GitOpsRepoURL: "https://example.com/testing/test-repo.git",
	})
	assertNoError(t, err)
	refuteRepositoryCreated(t, fakeData)
}

func TestRepoURL(t *testing.T) {
	urlTests := []struct {
		repoURL string
		wantURL string
	}{
		{"https://github.com/my-org/my-repo.git", "https://github.com"},
		{"https://gl.example.com/my-org/my-repo.git", "https://gl.example.com"},
	}

	for _, tt := range urlTests {
		t.Run(tt.repoURL, func(rt *testing.T) {
			u, err := repoURL(tt.repoURL)
			if err != nil {
				rt.Error(err)
				return
			}
			if u != tt.wantURL {
				rt.Errorf("got %q, want %q", u, tt.wantURL)
			}
		})
	}
}

func stubOutGitClientFactory(t *testing.T, authToken string) *fake.Data {
	t.Helper()
	f := defaultClientFactory
	t.Cleanup(func() {
		defaultClientFactory = f
	})

	client, data := fake.NewDefault()
	defaultClientFactory = func(repoURL string) (*scm.Client, error) {
		t.Helper()
		u, err := url.Parse(repoURL)
		if err != nil {
			return nil, err
		}
		want := ":" + authToken
		if a := u.User.String(); a != want {
			t.Fatalf("client failed auth: got %q, want %q", a, want)
		}
		return client, nil
	}
	return data
}

func assertRepositoryCreated(t *testing.T, data *fake.Data, org, name string) {
	t.Helper()
	want := []*scm.RepositoryInput{
		{
			Namespace:   org,
			Name:        name,
			Description: defaultRepoDescription,
			Private:     true,
		},
	}
	if diff := cmp.Diff(want, data.CreateRepositories); diff != "" {
		t.Fatalf("BootstrapRepository failed:\n%s", diff)
	}
}

func refuteRepositoryCreated(t *testing.T, data *fake.Data) {
	t.Helper()
	if l := len(data.CreateRepositories); l != 0 {
		t.Fatalf("BootstrapRepository created repositories: %d", l)
	}
}

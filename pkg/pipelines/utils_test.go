package pipelines

import (
	"bytes"
	"errors"
	"net/url"
	"sync"
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

func TestPushRepository(t *testing.T) {
	repo := "git@github.com:testing/testing.git"
	opts := &BootstrapOptions{
		OutputPath: "/tmp",
	}
	outputs := [][]byte{
		[]byte("Initialized empty Git repository in /tmp/.git/"),
		[]byte(""),
	}
	e := stubOutCmdExecution(t, outputs...)

	err := pushRepository(opts, repo)
	assertNoError(t, err)

	want := []execution{
		{
			BaseDir: opts.OutputPath,
			Command: "git",
			Args:    []string{"init", "."},
		},
		{
			BaseDir: opts.OutputPath,
			Command: "git",
			Args:    []string{"add", "."},
		},
		{
			BaseDir: opts.OutputPath,
			Command: "git",
			Args:    []string{"commit", "-m", "Bootstrapped commit"},
		},
		{
			BaseDir: opts.OutputPath,
			Command: "git",
			Args:    []string{"branch", "-m", "main"},
		},
		{
			BaseDir: opts.OutputPath,
			Command: "git",
			Args:    []string{"remote", "add", "origin", repo},
		},
		{
			BaseDir: opts.OutputPath,
			Command: "git",
			Args:    []string{"push", "-u", "origin", "main"},
		},
	}
	if diff := cmp.Diff(want, e.executions()); diff != "" {
		t.Fatalf("failed to push the repository:\n%s", diff)
	}
}

func TestPushRepository_handling_errors(t *testing.T) {
	repo := "git@github.com:testing/testing.git"
	opts := &BootstrapOptions{
		OutputPath: "/tmp",
	}
	outputs := [][]byte{
		[]byte("failed to create /tmp/.git/"),
	}
	testErr := errors.New("test error")
	e := stubOutCmdExecution(t, outputs...)
	e.errors.push(nil)
	e.errors.push(testErr)

	err := pushRepository(opts, repo)
	if !errors.Is(err, testErr) {
		t.Fatalf("got error %v, want %v", err, testErr)
	}

	want := []execution{
		{
			BaseDir: opts.OutputPath,
			Command: "git",
			Args:    []string{"init", "."},
		},
	}
	if diff := cmp.Diff(want, e.executions()); diff != "" {
		t.Fatalf("failed to push the repository:\n%s", diff)
	}
}

func TestCmdExecutor(t *testing.T) {
	var e executor = cmdExecutor{}
	out, err := e.execute(".", "git status", "")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(out, []byte("On branch")) {
		t.Fatalf("didn't get git output: %s", out)
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

func stubOutCmdExecution(t *testing.T, outputs ...[]byte) *mockExecutor {
	t.Helper()
	f := defaultExecutor
	t.Cleanup(func() {
		defaultExecutor = f
	})
	e := &mockExecutor{
		outputs:  newOutputs(outputs...),
		errors:   newErrors(),
		executed: []execution{},
	}
	defaultExecutor = e
	return e
}

type mockExecutor struct {
	outputs  *outputStack
	errors   *errorStack
	executed []execution
}

type execution struct {
	BaseDir string
	Command string
	Args    []string
}

func (m *mockExecutor) execute(basedir, command string, args ...string) ([]byte, error) {
	m.executed = append(m.executed, execution{BaseDir: basedir, Command: command, Args: args})
	return m.outputs.pop(), m.errors.pop()
}

func (m *mockExecutor) executions() []execution {
	return m.executed
}

type errorStack struct {
	errors []error
	sync.Mutex
}

func newErrors() *errorStack {
	return &errorStack{
		errors: []error{},
	}
}

func (s *errorStack) push(err error) {
	s.Lock()
	defer s.Unlock()
	s.errors = append(s.errors, err)
}

func (s *errorStack) pop() error {
	s.Lock()
	defer s.Unlock()
	if len(s.errors) == 0 {
		return nil
	}
	err := s.errors[len(s.errors)-1]
	s.errors = s.errors[0 : len(s.errors)-1]
	return err
}

type outputStack struct {
	outputs [][]byte
	sync.Mutex
}

func newOutputs(o ...[]byte) *outputStack {
	return &outputStack{
		outputs: o,
	}
}

func (s *outputStack) pop() []byte {
	s.Lock()
	defer s.Unlock()
	if len(s.outputs) == 0 {
		return []byte("")
	}
	o := s.outputs[len(s.outputs)-1]
	s.outputs = s.outputs[0 : len(s.outputs)-1]
	return o
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

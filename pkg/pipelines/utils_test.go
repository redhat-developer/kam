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
	"github.com/redhat-developer/kam/test"
)

func TestBootstrapRepository_with_personal_account(t *testing.T) {
	token := "this-is-a-test-token"
	factory, fakeData := newMockClientFactory(t, token)
	fakeData.CurrentUser = scm.User{Login: "testing"}

	err := BootstrapRepository(
		&BootstrapOptions{
			GitOpsRepoURL:      "https://example.com/testing/test-repo.git",
			GitHostAccessToken: token,
		},
		factory,
		newMockExecutor(),
	)
	assertNoError(t, err)

	assertRepositoryCreated(t, fakeData, "", "test-repo")
}

func TestBootstrapRepository_with_org(t *testing.T) {
	token := "this-is-a-test-token"
	factory, fakeData := newMockClientFactory(t, token)
	fakeData.CurrentUser = scm.User{Login: "test-user"}

	err := BootstrapRepository(
		&BootstrapOptions{
			GitOpsRepoURL:      "https://example.com/testing/test-repo.git",
			GitHostAccessToken: token,
		},
		factory,
		newMockExecutor(),
	)
	assertNoError(t, err)
	assertRepositoryCreated(t, fakeData, "testing", "test-repo")
}

func TestBootstrapRepository_with_no_access_token(t *testing.T) {
	token := "this-is-a-test-token"
	factory, fakeData := newMockClientFactory(t, token)
	fakeData.CurrentUser = scm.User{Login: "test-user"}

	err := BootstrapRepository(
		&BootstrapOptions{
			GitOpsRepoURL: "https://example.com/testing/test-repo.git",
		},
		factory,
		newMockExecutor(),
	)
	assertNoError(t, err)
	refuteRepositoryCreated(t, fakeData)
}

func TestPushRepositoryWithSetURL(t *testing.T) {
	repo := "git@github.com:testing/testing.git"
	opts := &BootstrapOptions{
		OutputPath: "/tmp",
	}
	outputs := [][]byte{
		[]byte("Initialized empty Git repository in /tmp/.git/"),
		[]byte(""),
	}
	e := newMockExecutor(outputs...)

	err := pushRepository(opts, repo, e)
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
			Args:    []string{"add", "pipelines.yaml", "config", "environments"},
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
			Args:    []string{"remote", "show", "origin"},
		},
		{
			BaseDir: opts.OutputPath,
			Command: "git",
			Args:    []string{"remote", "set-url", "origin", repo},
		},
		{
			BaseDir: opts.OutputPath,
			Command: "git",
			Args:    []string{"push", "-u", "origin", "main"},
		},
	}
	e.assertCommandsExecuted(t, want)
}

func TestPushRepositoryWithRemoteAdd(t *testing.T) {
	repo := "git@github.com:testing/testing.git"
	opts := &BootstrapOptions{
		OutputPath: "/tmp",
	}
	outputs := [][]byte{
		[]byte("Initialized empty Git repository in /tmp/.git/"),
		[]byte(""),
	}
	e := newMockExecutor(outputs...)
	testErr := errors.New("test error")
	e.errors.push(testErr)
	e.errors.push(nil)
	e.errors.push(nil)
	e.errors.push(nil)
	e.errors.push(nil)
	err := pushRepository(opts, repo, e)
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
			Args:    []string{"add", "pipelines.yaml", "config", "environments"},
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
			Args:    []string{"remote", "show", "origin"},
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
	e.assertCommandsExecuted(t, want)
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
	e := newMockExecutor(outputs...)
	e.errors.push(nil)
	e.errors.push(testErr)

	err := pushRepository(opts, repo, e)
	test.AssertErrorMatch(t, "test error", err)

	want := []execution{
		{
			BaseDir: opts.OutputPath,
			Command: "git",
			Args:    []string{"init", "."},
		},
	}
	e.assertCommandsExecuted(t, want)
}

func TestCmdExecutor(t *testing.T) {
	var e executor = cmdExecutor{}
	out, err := e.execute(".", "echo", "hello")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(out, []byte("hello")) {
		t.Fatalf("didn't get git output: %s", out)
	}
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

func newMockClientFactory(t *testing.T, authToken string) (clientFactory, *fake.Data) {
	client, data := fake.NewDefault()
	f := func(repoURL string) (*scm.Client, error) {
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
	return f, data
}

func newMockExecutor(outputs ...[]byte) *mockExecutor {
	return &mockExecutor{
		outputs:  newOutputs(outputs...),
		errors:   newErrors(),
		executed: []execution{},
	}
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

func (m *mockExecutor) assertCommandsExecuted(t *testing.T, want []execution) {
	if diff := cmp.Diff(want, m.executed); diff != "" {
		t.Fatalf("failed to push the repository:\n%s", diff)
	}
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

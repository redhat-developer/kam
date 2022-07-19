package bootstrapnew

import (
	"regexp"
	"testing"

	pipelines "github.com/redhat-developer/kam/pkg/pipelines/component"
)

// Test Component Parameters

func TestValidateBootstrapNewParameter(t *testing.T) {
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
		o := BootstrapNewParameters{
			BootstrapNewOptions: &pipelines.BootstrapNewOptions{
				GitRepoURL:           tt.gitRepo,
				PrivateRepoURLDriver: tt.driver,
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

// Test case for checking the input flags.
func TestMissingFlagsBootstrapNew(t *testing.T) {
	tests := []struct {
		desc  string
		flags map[string]string
		err   error
	}{
		{
			"Required flags are present",
			map[string]string{"component-name": "value-1", "application-name": "value-2", "git-repo-url": "value-3"},
			nil,
		},
		{
			"A required flag is absent",
			map[string]string{"component-name": "", "application-name": "value-2", "git-repo-url": "value-3"},
			missingFlagErr([]string{`"component-name"`}),
		},
		{
			"A required flag is absent",
			map[string]string{"component-name": "value-1", "application-name": "", "git-repo-url": "value-3"},
			missingFlagErr([]string{`"application-name"`}),
		},
		{
			"A required flag is absent",
			map[string]string{"component-name": "value-1", "application-name": "value-2", "git-repo-url": ""},
			missingFlagErr([]string{`"git-repo-url"`}),
		},
		{
			"Multiple required flags are absent",
			map[string]string{"component-name": "", "application-name": "", "git-repo-url": ""},
			missingFlagErr([]string{`"component-name"`, `"application-name"`, `"git-repo-url"`}),
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

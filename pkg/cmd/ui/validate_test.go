package ui

import (
	"testing"
)

func TestValidatePrefix(t *testing.T) {

	validator := makePrefixValidator()
	cmdTests := []struct {
		desc     string
		argument string
		wantErr  string
	}{
		{"Name is not valid",
			"Test@",
			`Test@-stage is not a valid name:  a DNS-1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')`},
		{"Prefix too long",
			"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz",
			"The prefix abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz-, must be less than 58 characters",
		},
	}

	for _, tt := range cmdTests {
		t.Run(tt.desc, func(t *testing.T) {
			err := validator(tt.argument)
			if err.Error() != tt.wantErr {
				t.Errorf("got %s, want %s", err, tt.wantErr)
			}
		})
	}
}

func TestValidateSecretLength(t *testing.T) {
	validator := makeSecretValidator()
	cmdTests := []struct {
		desc     string
		argument string
		wantErr  string
	}{
		{"Secret length too short",
			"abc",
			`The length of the secret must be at least 16 characters`},
	}

	for _, tt := range cmdTests {
		t.Run(tt.desc, func(t *testing.T) {
			err := validator(tt.argument)
			if err.Error() != tt.wantErr {
				t.Errorf("got %s, want %s", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAccessToken(t *testing.T) {
	cmdTests := []struct {
		desc     string
		argument *RepoParams
		wantErr  string
	}{
		{"Invalid Driver",
			&RepoParams{TokenRepoMatchCondition: false, RepoInfo: repoInfo{
				RepoURL: "https://test.com/username/repo.git",
			}},
			`unable to identify driver from hostname: test.com`},
		{"Invalid Access Token",
			&RepoParams{TokenRepoMatchCondition: false, GitHostAccessToken: "abc123", RepoInfo: repoInfo{
				RepoURL: "https://github.com/username/repo.git",
			}},
			`forbidden: Invalid access token, unable to authenticate client for repo: https://github.com/username/repo.git`},
		{"Invalid Repo-Name",
			&RepoParams{TokenRepoMatchCondition: false, GitHostAccessToken: "abc123", RepoInfo: repoInfo{
				RepoURL: "https://github.com/username./repo.git",
			}},
			`unable to get the repo name from "https://github.com/username./repo.git": failed to get Git repo: /username./repo.git`},
	}

	for _, tt := range cmdTests {
		t.Run(tt.desc, func(t *testing.T) {
			err := ValidateAccessToken(tt.argument)
			if err.Error() != tt.wantErr {
				t.Errorf("got %s, want %s", err, tt.wantErr)
			}
		})
	}
}

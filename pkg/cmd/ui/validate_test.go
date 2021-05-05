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
			`Test@-stage is not a valid name:  a lowercase RFC 1123 label must consist of lower case alphanumeric characters or '-', and must start and end with an alphanumeric character (e.g. 'my-name',  or '123-abc', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?')`},
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

func TestAccessToken(t *testing.T) {
	mockurl := "https://github.com/example/test.git"
	validator := makeAccessTokenCheck(mockurl)
	cmdTests := []struct {
		desc     string
		argument string
		wantErr  string
	}{
		{"Access Token is incorrect",
			"demo-token",
			`The token passed is incorrect for repository example/test`},
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

func TestValidateURL(t *testing.T) {

	validator := makeURLValidatorCheck()
	cmdTests := []struct {
		desc     string
		argument string
		wantErr  string
	}{
		{
			"Invalid URL format",
			"gitops repo https://github.com/test/gitops.git",
			"invalid URL, err: parse \"gitops repo https://github.com/test/gitops.git\": first path segment in URL cannot contain colon",
		},
		{
			"Empty URL",
			"",
			"could not identify host from \"\"",
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

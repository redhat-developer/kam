package ui

import (
	"testing"
)

func TestValidateSecretLength(t *testing.T) {
	validator := makeSecretValidator()
	cmdTests := []struct {
		desc     string
		argument string
		wantErr  string
	}{
		{"Secret length too short",
			"abc",
			`the length of the secret must be at least 16 characters`},
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

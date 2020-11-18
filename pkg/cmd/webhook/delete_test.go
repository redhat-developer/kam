package webhook

import (
	"fmt"
	"testing"
)

func TestValidateForDelete(t *testing.T) {

	testcases := []struct {
		options *deleteOptions
		errMsg  string
	}{
		{
			&deleteOptions{
				options{isCICD: true, serviceName: "foo"},
			},
			"Only one of 'cicd' or 'env-name/service-name' can be specified",
		},
		{
			&deleteOptions{
				options{isCICD: true, envName: "foo"},
			},
			"Only one of 'cicd' or 'env-name/service-name' can be specified",
		},
		{
			&deleteOptions{
				options{isCICD: true, envName: "foo", serviceName: "bar"},
			},
			"Only one of 'cicd' or 'env-name/service-name' can be specified",
		},
		{
			&deleteOptions{
				options{isCICD: false},
			},
			"One of 'cicd' or 'env-name/service-name' must be specified",
		},
		{
			&deleteOptions{
				options{isCICD: false, serviceName: "foo"},
			},
			"One of 'cicd' or 'env-name/service-name' must be specified",
		},
		{
			&deleteOptions{
				options{isCICD: false, serviceName: "foo", envName: "gau"},
			},
			"",
		},
		{
			&deleteOptions{
				options{isCICD: true, serviceName: ""},
			},
			"",
		},
	}

	for i, tt := range testcases {
		t.Run(fmt.Sprintf("Test %d", i), func(t *testing.T) {
			err := tt.options.Validate()
			if err != nil && tt.errMsg == "" {
				t.Errorf("Validate() got an unexpected error: %s", err)
			} else {
				if !matchError(t, tt.errMsg, err) {
					t.Errorf("Validate() failed to match error: got %s, want %s", err, tt.errMsg)
				}
			}
		})
	}
}

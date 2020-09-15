package pipelines

import "testing"

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

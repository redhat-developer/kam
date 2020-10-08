package git

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
	"github.com/jenkins-x/go-scm/scm/factory"
)

var mockHeaders = map[string]string{
	"X-GitHub-Request-Id":   "DD0E:6011:12F21A8:1926790:5A2064E2",
	"X-RateLimit-Limit":     "60",
	"X-RateLimit-Remaining": "59",
	"X-RateLimit-Reset":     "1512076018",
}

func TestWebhookWithFakeClient(t *testing.T) {
	fakeID := factory.NewDriverIdentifier(factory.Mapping("fake.com", "fake"))
	factory.DefaultIdentifier = fakeID
	repo, err := NewRepository("https://fake.com/foo/bar.git", "token")
	if err != nil {
		t.Fatal(err)
	}

	listenerURL := "http://example.com/webhook"
	ids, err := repo.ListWebhooks(listenerURL)
	if err != nil {
		t.Fatal(err)
	}

	// start with no webhooks
	if len(ids) > 0 {
		t.Fatalf("got %d ids, want 0", len(ids))
	}

	// create a webhook
	id, err := repo.CreateWebhook(listenerURL, "secret")
	if err != nil {
		t.Fatal(err)
	}

	// verify and remember our ID
	if id == "" {
		t.Fatal("got no webhook id")
	}

	// list again
	ids, err = repo.ListWebhooks(listenerURL)
	if err != nil {
		t.Fatal(err)
	}

	// verify ID from list
	if diff := cmp.Diff(ids, []string{id}); diff != "" {
		t.Fatalf("created id mismatch got\n%s", diff)
	}

	// delete webhook
	deleted, err := repo.DeleteWebhooks(ids)
	if err != nil {
		t.Fatal(err)
	}

	// verify deleted IDs
	if diff := cmp.Diff(ids, deleted); diff != "" {
		t.Fatalf("deleted ids mismatch got\n%s", diff)
	}

	ids, err = repo.ListWebhooks(listenerURL)
	if err != nil {
		t.Fatal(err)
	}

	// verify no webhooks
	if len(ids) > 0 {
		t.Fatalf("got %d ids, want 0", len(ids))
	}
}

func TestListWebHooks(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Get("/repos/foo/bar/hooks").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/hooks.json")

	repo, err := NewRepository("https://github.com/foo/bar.git", "token")
	if err != nil {
		t.Fatal(err)
	}

	ids, err := repo.ListWebhooks("http://example.com/webhook")
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(ids, []string{"1"}); diff != "" {
		t.Errorf("driver errMsg mismatch got\n%s", diff)
	}
}

func TestDeleteWebHooks(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Delete("/repos/foo/bar/hooks/1").
		Reply(204).
		Type("application/json").
		SetHeaders(mockHeaders)

	repo, err := NewRepository("https://github.com/foo/bar.git", "token")
	if err != nil {
		t.Fatal(err)
	}

	deleted, err := repo.DeleteWebhooks([]string{"1"})
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff([]string{"1"}, deleted); diff != "" {
		t.Errorf("deleted mismatch got\n%s", diff)
	}
}

func TestCreateWebHook(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.github.com").
		Post("/repos/foo/bar/hooks").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/hook.json")

	repo, err := NewRepository("https://github.com/foo/bar.git", "token")
	if err != nil {
		t.Fatal(err)
	}

	created, err := repo.CreateWebhook("http://example.com/webhook", "mysecret")
	if err != nil {
		t.Fatal(err)
	}

	if created != "1" {
		t.Errorf("failed to create webhook, got %q, want %q", created, "1")
	}
}

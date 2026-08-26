package issues

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v66/github"
)

// TestGitHubServiceCreate verifies that Create sends a correctly formed POST request to GitHub's issues endpoint and correctly translates a successful response back into this package's Issue type.
func TestGitHubServiceCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/repos/vmkalima/repo/issues"
		if r.URL.Path != wantPath {
			t.Errorf("expected request to %q, got %q", wantPath, r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		var body github.IssueRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.Title == nil || *body.Title != "New issue" {
			t.Errorf("expected title %q, got %v", "New issue", body.Title)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if _, err := fmt.Fprint(w, `{"number": 42, "title": "New issue", "state": "open"}`); err != nil {
			t.Fatalf("failed to write test response: %v", err)
		}
	}))
	defer server.Close()

	service := newTestGitHubService(t, server.URL)

	issue, err := service.Create(context.Background(), "vmkalima", "repo", "New issue", "")
	if err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}

	if issue.Number != 42 {
		t.Errorf("expected issue number 42, got %d", issue.Number)
	}
	if issue.Title != "New issue" {
		t.Errorf("expected title %q, got %q", "New issue", issue.Title)
	}
	if issue.State != "open" {
		t.Errorf("expected state %q, got %q", "open", issue.State)
	}
}

// TestGitHubServiceCreateError verifies that Create returns a non-nil error when GitHub responds with a validation failure, rather than silently succeeding or panicking on an unexpected response shape.
func TestGitHubServiceCreateError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		if _, err := fmt.Fprint(w, `{"message": "Validation Failed"}`); err != nil {
			t.Fatalf("failed to write test response: %v", err)
		}
	}))
	defer server.Close()

	service := newTestGitHubService(t, server.URL)

	_, err := service.Create(context.Background(), "vmkalima", "repo", "New issue", "")
	if err == nil {
		t.Fatal("expected an error when GitHub returns a failure status, got nil")
	}
}

// TestGitHubServiceList verifies that List sends a correctly formed GET request and correctly translates a multi-issue response, preserving order, into a slice of Issue.
func TestGitHubServiceList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/repos/vmkalima/repo/issues"
		if r.URL.Path != wantPath {
			t.Errorf("expected request to %q, got %q", wantPath, r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprint(w, `[
			{"number": 1, "title": "First", "state": "open"},
			{"number": 2, "title": "Second", "state": "closed"}
		]`); err != nil {
			t.Fatalf("failed to write test response: %v", err)
		}
	}))
	defer server.Close()

	service := newTestGitHubService(t, server.URL)

	got, err := service.List(context.Background(), "vmkalima", "repo")
	if err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 issues, got %d", len(got))
	}
	if got[0].Number != 1 || got[0].Title != "First" || got[0].State != "open" {
		t.Errorf("unexpected first issue: %+v", got[0])
	}
	if got[1].Number != 2 || got[1].Title != "Second" || got[1].State != "closed" {
		t.Errorf("unexpected second issue: %+v", got[1])
	}
}

// TestGitHubServiceListEmpty verifies that List correctly handles a repository with no issues, returning an empty (not nil-panicking) slice rather than erroring on an empty JSON array.
func TestGitHubServiceListEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprint(w, `[]`); err != nil {
			t.Fatalf("failed to write test response: %v", err)
		}
	}))
	defer server.Close()

	service := newTestGitHubService(t, server.URL)

	got, err := service.List(context.Background(), "vmkalima", "repo")
	if err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected 0 issues, got %d", len(got))
	}
}

// TestGitHubServiceListError verifies that List returns a non-nil error when GitHub responds with a failure status.
func TestGitHubServiceListError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	service := newTestGitHubService(t, server.URL)

	_, err := service.List(context.Background(), "vmkalima", "repo")
	if err == nil {
		t.Fatal("expected an error when GitHub returns a failure status, got nil")
	}
}

// TestGitHubServiceClose verifies that Close sends a PATCH request setting the issue's state to "closed", and correctly translates the resulting response back into this package's Issue type.
func TestGitHubServiceClose(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wantPath := "/repos/vmkalima/repo/issues/7"
		if r.URL.Path != wantPath {
			t.Errorf("expected request to %q, got %q", wantPath, r.URL.Path)
		}
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}

		var body github.IssueRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.State == nil || *body.State != "closed" {
			t.Errorf("expected state %q, got %v", "closed", body.State)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprint(w, `{"number": 7, "title": "To close", "state": "closed"}`); err != nil {
			t.Fatalf("failed to write test response: %v", err)
		}
	}))
	defer server.Close()

	service := newTestGitHubService(t, server.URL)

	issue, err := service.Close(context.Background(), "vmkalima", "repo", 7)
	if err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}

	if issue.Number != 7 {
		t.Errorf("expected issue number 7, got %d", issue.Number)
	}
	if issue.State != "closed" {
		t.Errorf("expected state %q, got %q", "closed", issue.State)
	}
}

// TestGitHubServiceCloseNotFound verifies that Close returns a non-nil error when GitHub reports the issue does not exist (404), which is the most likely real-world failure mode for this operation.
func TestGitHubServiceCloseNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if _, err := fmt.Fprint(w, `{"message": "Not Found"}`); err != nil {
			t.Fatalf("failed to write test response: %v", err)
		}
	}))
	defer server.Close()

	service := newTestGitHubService(t, server.URL)

	_, err := service.Close(context.Background(), "vmkalima", "repo", 999)
	if err == nil {
		t.Fatal("expected an error when closing a nonexistent issue, got nil")
	}
}

// newTestGitHubService builds a GitHubService whose client talks to the given test server URL instead of the real GitHub API, so tests can exercise GitHubService's request construction and response parsing without depending on network access or a real GitHub token.
func newTestGitHubService(t *testing.T, serverURL string) *GitHubService {
	t.Helper()

	client := github.NewClient(nil)
	baseURL, err := url.Parse(serverURL + "/")
	if err != nil {
		t.Fatalf("failed to parse test server URL: %v", err)
	}
	client.BaseURL = baseURL

	return &GitHubService{client: client}
}
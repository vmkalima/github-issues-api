package api

import(
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmkalima/github-issues-api/internal/issues"
)

// TestCreateIssue tests the CreateIssue handler of the API. It verifies that a new issue can be created successfully and that the response contains the expected issue details.
func TestCreateIssue(t *testing.T) {
	handler := NewHandler(issues.NewFake())

	reqBody := bytes.NewBufferString(`{"title": "Test Issue"}`)
	req := httptest.NewRequest(http.MethodPost, "/repos/testowner/testrepo/issues", reqBody)

	// Since the request is built directly and not muxed, we need to set the path parameters manually. In a real scenario, these would be extracted from the URL by the router.
	req.SetPathValue("owner", "testowner")
	req.SetPathValue("repo", "testrepo")
	w := httptest.NewRecorder()

	handler.CreateIssue(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var issue issues.Issue
	if err := json.NewDecoder(w.Body).Decode(&issue); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if issue.Title != "Test Issue" {
		t.Errorf("Expected issue title 'Test Issue', got '%s'", issue.Title)
	}

	if issue.State != "open" {
		t.Errorf("Expected issue state 'open', got '%s'", issue.State)
	}

	if issue.Number != 1 {
		t.Errorf("Expected issue number 1, got %d", issue.Number)
	}
}

// TestCreateIssueEmptyTitle tests the CreateIssue handler with an empty title. It verifies that the handler responds with a Bad Request status code when the title is missing.
func TestCreateIssueEmptyTitle(t *testing.T) {
	handler := NewHandler(issues.NewFake())

	reqbody := bytes.NewBufferString(`{"title": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/repos/testowner/testrepo/issues", reqbody)
	w := httptest.NewRecorder()

	handler.CreateIssue(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestCreateIssueLongTitle tests the CreateIssue handler with a title that exceeds the maximum allowed length. It verifies that the handler responds with a Bad Request status code when the title is too long.
func TestCreateIssueLongTitle(t *testing.T) {
	handler := NewHandler(issues.NewFake())

	longTitle := make([]byte, 257)
	for i := range longTitle {
		longTitle[i] = 'a'
	}
	reqbody := bytes.NewBufferString(`{"title": "` + string(longTitle) + `"}`)
	req := httptest.NewRequest(http.MethodPost, "/repos/testowner/testrepo/issues", reqbody)
	w := httptest.NewRecorder()

	handler.CreateIssue(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestCreateIssueInvalidJSON tests the CreateIssue handler with an invalid JSON request body. It verifies that the handler responds with a Bad Request status code when the request body cannot be parsed as valid JSON.
func TestCreateIssueInvalidJSON(t *testing.T) {
	handler := NewHandler(issues.NewFake())

	reqbody := bytes.NewBufferString(`invalid json`)
	req := httptest.NewRequest(http.MethodPost, "/repos/testowner/testrepo/issues", reqbody)
	w := httptest.NewRecorder()

	handler.CreateIssue(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}
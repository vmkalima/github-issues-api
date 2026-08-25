package api

import(
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"context"
	"strconv"

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

// TestListIssues tests the ListIssues handler of the API. It verifies that the handler correctly retrieves and returns a list of issues for a specified repository.
func TestListIssues(t *testing.T) {
	service := issues.NewFake()
	handler := NewHandler(service)

	ctx := context.Background()
	if _, err := service.Create(ctx, "testowner", "testrepo", "Issue 1", ""); err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}
	if _, err := service.Create(ctx, "testowner", "testrepo", "Issue 2", ""); err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/repos/testowner/testrepo/issues", nil)
	req.SetPathValue("owner", "testowner")
	req.SetPathValue("repo", "testrepo")
	w := httptest.NewRecorder()

	handler.ListIssues(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
	
	var issuesList []issues.Issue
	if err := json.NewDecoder(w.Body).Decode(&issuesList); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(issuesList) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(issuesList))
	}
}

// TestListIssuesEmpty tests the ListIssues handler when there are no issues in the specified repository. It verifies that the handler returns an empty list and a 200 OK status code.
func TestListIssuesEmpty(t *testing.T) {
	service := issues.NewFake()
	handler := NewHandler(service)

	req := httptest.NewRequest(http.MethodGet, "/repos/testowner/testrepo/issues", nil)
	req.SetPathValue("owner", "testowner")
	req.SetPathValue("repo", "testrepo")
	w := httptest.NewRecorder()

	handler.ListIssues(w, req)
	
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var issuesList []issues.Issue
	if err := json.NewDecoder(w.Body).Decode(&issuesList); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if len(issuesList) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issuesList))
	}
}

// TestCloseIssue tests the CloseIssue handler of the API. It verifies that an existing issue can be closed successfully and that the response reflects the updated state of the issue.
func TestCloseIssue(t *testing.T) {
	service := issues.NewFake()
	handler := NewHandler(service)

	ctx := context.Background()
	issue, err := service.Create(ctx, "testowner", "testrepo", "Issue to close", "")
	if err != nil {
		t.Fatalf("Failed to create issue: %v", err)
	}

	req := httptest.NewRequest(http.MethodDelete, "/repos/testowner/testrepo/issues/1", nil)
	req.SetPathValue("owner", "testowner")
	req.SetPathValue("repo", "testrepo")
	req.SetPathValue("number", strconv.Itoa(issue.Number))
	w := httptest.NewRecorder()

	handler.CloseIssue(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	var closedIssue issues.Issue
	if err := json.NewDecoder(w.Body).Decode(&closedIssue); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if closedIssue.State != "closed" {
		t.Errorf("Expected issue state 'closed', got '%s'", closedIssue.State)
	}
}

// TestCloseIssueNotFound tests the CloseIssue handler when attempting to close a non-existent issue. It verifies that the handler responds with a Not Found status code.
func TestCloseIssueNotFound(t *testing.T) {
	handler := NewHandler(issues.NewFake())

	req := httptest.NewRequest(http.MethodDelete, "/repos/testowner/testrepo/issues/42", nil)
	req.SetPathValue("owner", "testowner")
	req.SetPathValue("repo", "testrepo")
	req.SetPathValue("number", "42")
	w := httptest.NewRecorder()

	handler.CloseIssue(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("Expected status code %d, got %d", http.StatusNotFound, w.Code)
	}
}

// TestCloseIssueInvalidNumber tests the CloseIssue handler with an invalid issue number. It verifies that the handler responds with a Bad Request status code when the issue number is not a valid integer.
func TestCloseIssueInvalidNumber(t *testing.T) {
	handler := NewHandler(issues.NewFake())

	req := httptest.NewRequest(http.MethodDelete, "/repos/testowner/testrepo/issues/invalid", nil)
	req.SetPathValue("owner", "testowner")
	req.SetPathValue("repo", "testrepo")
	req.SetPathValue("number", "invalid")
	w := httptest.NewRecorder()

	handler.CloseIssue(w, req)
	
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status code %d, got %d", http.StatusBadRequest, w.Code)
	}
}
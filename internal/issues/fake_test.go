package issues

import (
	"context"
	"testing"
)

// TestFakeCreateAndList tests the Create and List methods of the Fake service, ensuring that issues can be created and retrieved correctly.
func TestFakeCreateAndList(t *testing.T) {
	ctx := context.Background()
	fake := NewFake()

	created, err := fake.Create(ctx, "vmkalima", "some-repo", "First issue")
	if err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}
	if (created.State != "open") {
		t.Errorf("Expected issue state to be 'open', got '%s'", created.State)
	}

	issues, err := fake.List(ctx, "vmkalima", "some-repo")
	if err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Errorf("Expected 1 issue, got %d", len(issues))
	}
}

// TestFakeClose tests the Close method of the Fake service, ensuring that an issue can be closed and its state is updated accordingly.
func TestFakeClose(t *testing.T) {
	ctx := context.Background()
	fake := NewFake()

	created, err := fake.Create(ctx, "vmkalima", "some-repo", "To be closed")
	if err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}

	closed, err := fake.Close(ctx, "vmkalima", "some-repo", created.Number)
	if err != nil {
		t.Fatalf("Close returned unexpected error: %v", err)
	}
	if (closed.State != "closed") {
		t.Errorf("Expected issue state to be 'closed', got '%s'", closed.State)
	}
}

// TestFakeCloseNotFound tests the Close method of the Fake service when attempting to close a non-existent issue, ensuring that an appropriate error is returned.
func TestFakeCloseNotFound(t *testing.T) {
	ctx := context.Background()
	fake := NewFake()

	_, err := fake.Close(ctx, "vmkalima", "some-repo", 42)
	if err == nil {
		t.Fatalf("Expected error when closing non-existent issue, got nil")
	}
}

// TestFakeListEmpty tests the List method of the Fake service when no issues have been created, ensuring that it returns an empty slice without errors.
func TestFakeListEmpty(t *testing.T) {
	ctx := context.Background()
	fake := NewFake()

	issues, err := fake.List(ctx, "vmkalima", "some-repo")
	if err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(issues))
	}
}

// TestFakeRepoIsolation tests that issues created in different repositories are isolated and do not interfere with each other, ensuring that the Fake service maintains separate state for each repository.
func TestFakeRepoIsolation(t *testing.T) {
	ctx := context.Background()
	fake := NewFake()

	issueA, err := fake.Create(ctx, "vmkalima", "repoA", "Issue in repo A")
	if err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}

	issueB, err := fake.Create(ctx, "vmkalima", "repoB", "Issue in repo B")
	if err != nil {
		t.Fatalf("Create returned unexpected error: %v", err)
	}

	if issueA.Number != 1 || issueB.Number != 1 {
		t.Errorf("Expected both issues to have number 1, got %d and %d", issueA.Number, issueB.Number)
	}
	
	issuesA, err := fake.List(ctx, "vmkalima", "repoA")
	if err != nil {
		t.Fatalf("List returned unexpected error: %v", err)
	}
	issuesB, err_ := fake.List(ctx, "vmkalima", "repoB")
	if err_ != nil {
		t.Fatalf("List returned unexpected error: %v", err_)
	}

	if len(issuesA) != 1 || len(issuesB) != 1 {
		t.Fatalf("Expected 1 issue in each repo, got %d in repoA and %d in repoB", len(issuesA), len(issuesB))
	}
}
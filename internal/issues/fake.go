package issues

import (
	"context"
	"fmt"
	"sync"
)

var _ Service = (*Fake)(nil) // Ensure Fake implements the Service interface

// Fake is an in-memory implementation of the Service interface, used for testing purposes. It simulates the behavior of a real GitHub issues service without making actual API calls.
type Fake struct {
	mu	 sync.Mutex			// mutual exclusion lock--prevent simultaneous access 
	nextID map[string]int
	issues map[string]map[int]*Issue
}

// NewFake creates and returns a new instance of the Fake service, initializing its internal state.
func NewFake() *Fake {
	return &Fake{
		nextID: make(map[string]int),
		issues: make(map[string]map[int]*Issue),
	}
}

// repoKey generates a unique key for a given repository based on its owner and name, used for internal storage in the Fake service.
func repoKey(owner, repo string) string {
	return owner + "/" + repo
}

// Create creates a new issue in the specified repository with the given title. It returns the created issue or an error if the operation fails.
func (f *Fake) Create(ctx context.Context, owner, repo, title string) (*Issue, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := repoKey(owner, repo)

	if f.issues[key] == nil {
		f.issues[key] = make(map[int]*Issue)
	}

	f.nextID[key]++
	issue := &Issue{
		Number: f.nextID[key],
		Title:  title,
		State:  "open",
	}
	f.issues[key][issue.Number] = issue
	
	return issue, nil
}

// List retrieves all issues for the specified repository. It returns a slice of issues or an error if the operation fails.
func (f *Fake) List(ctx context.Context, owner, repo string) ([]*Issue, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := repoKey(owner, repo)
	repoIssues := f.issues[key]

	issues := make([]*Issue, 0, len(repoIssues))
	for _, issue := range repoIssues {
		issues = append(issues, issue)
	}
	return issues, nil
}

// Close closes the issue with the specified number in the given repository. It returns the closed issue or an error if the issue does not exist.
func (f *Fake) Close(ctx context.Context, owner, repo string, number int) (*Issue, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	key := repoKey(owner, repo)
	issue, ok := f.issues[key][number]
	if !ok {
		return nil, fmt.Errorf("issue #%d not found in %s", number, key)
	}
	issue.State = "closed"
	return issue, nil
}
package issues

import (
	"context"
	"fmt"

	"github.com/google/go-github/v66/github"
)

// GitHubService is a Service implementation backed by the real GitHub API
type GitHubService struct {
	client *github.Client
}

// NewGitHubService returns a GitHubService authenticated with the given GitHub token, scoped to whatever permissions that token carries.
func NewGithubService(token string) *GitHubService {
	client := github.NewClient(nil).WithAuthToken(token)
	return &GitHubService{client: client}
}

var _ Service = (*GitHubService)(nil)

// Create opens a new issue titled title in owner/repo via the GitHub API.
func (g *GitHubService) Create(ctx context.Context, owner, repo, title, body string) (*Issue, error) {
	ghIssue, _, err := g.client.Issues.Create(ctx, owner, repo, &github.IssueRequest{
		Title: &title,
		Body: &body,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating issue: %w", err)
	}
	return toIssue(ghIssue), nil
}

// List returns all issues currently open in owner/repo, as reported by the GitHub API.
func (g *GitHubService) List(ctx context.Context, owner, repo string) ([]*Issue, error) {
	ghIssues, _, err := g.client.Issues.ListByRepo(ctx, owner, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing issues: %w", err)
	}

	result := make([]*Issue, 0, len(ghIssues))
	for _, gi := range ghIssues {
		result = append(result, toIssue(gi))
	}
	return result, nil
}

// Close marks the issue numbered number in owner/repo as closed. GitHub's API does not support permanently deleting issues, so this is implemented as an edit setting the issue's state to "closed".
func (g *GitHubService) Close(ctx context.Context, owner, repo string, number int) (*Issue, error) {
	closedState := "closed"
	ghIssue, _, err := g.client.Issues.Edit(ctx, owner, repo, number, &github.IssueRequest{
		State: &closedState,
	})
	if err != nil {
		return nil, fmt.Errorf("closing issue %d: %w", number, err)
	}
	return toIssue(ghIssue), nil
}


// toIssue translates go-github's full Issue representation down to this package's minimal Issue type, exposing only the fields this API promises.
func toIssue(gi *github.Issue) *Issue {
	return &Issue{
		Number: gi.GetNumber(),
		Title: gi.GetTitle(),
		State: gi.GetState(),
		Body: gi.GetBody(),
	}
}
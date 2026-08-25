// Package issues defines the structure and interface for managing GitHub issues, independent of the underlying implementation (real GitHub API or a fake used for testing).
package issues

import "context"

// Issue represents a single GitHub issue as returned by the GitHub API
type Issue struct{
	Number int `json:"number"`
	Title string `json:"title"`
	State string `json:"state"`
	Body string `json:"body,omitempty"`
}

// Service defines the operations this API provides for managing issues on a given repository
type Service interface {
	Create(ctx context.Context, owner, repo, title string, body string) (*Issue, error)
	List(ctx context.Context, owner, repo string) ([]*Issue, error)
	Close(ctx context.Context, owner, repo string, number int) (*Issue, error)
}
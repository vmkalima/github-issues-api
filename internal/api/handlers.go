package api

import (
	"encoding/json"
	"net/http"
	"github.com/vmkalima/github-issues-api/internal/issues"
)

// maxTitleLength defines the maximum allowed length for an issue title. This constant is used to validate the title length when creating a new issue.
const maxTitleLength = 256

// Handler is a struct that holds the issues service, which is used to manage GitHub issues. It provides HTTP handlers for creating, listing, and closing issues.
type Handler struct {
	Service issues.Service
}

// NewHandler creates a new Handler instance with the provided issues service. It returns a pointer to the Handler, which can be used to handle HTTP requests related to GitHub issues.
func NewHandler(service issues.Service) *Handler {
	return &Handler{Service: service}
}

type createIssueRequest struct {
	Title string `json:"title"`
}

// CreateIssue handles the HTTP request for creating a new GitHub issue. It reads the request body to extract the issue title, calls the issues service to create the issue, and responds with the created issue in JSON format. If any error occurs during this process, it responds with an appropriate HTTP status code and error message.
func (h *Handler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")

	var req createIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	if len(req.Title) > maxTitleLength {
		http.Error(w, "Title exceeds maximum length of 256 characters", http.StatusBadRequest)
		return
	}

	issue, err := h.Service.Create(r.Context(), owner, repo, req.Title)
	if err != nil {
		http.Error(w, "Failed to create issue: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(issue); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
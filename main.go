package main

import (
    "log"
    "net/http"
    "os"
    "time"
    "github.com/vmkalima/github-issues-api/internal/issues"
    "github.com/vmkalima/github-issues-api/internal/api"
)

// healthHandler responds to health check requests with a simple "OK" body and a 200 status code.
func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    if _, err := w.Write([]byte("OK")); err != nil {
        log.Printf("Failed to write response: %v", err)
    }
}

func main() {

    apiToken := os.Getenv("API_TOKEN")
    if apiToken == "" {
        log.Fatal("API_TOKEN environment variable is not set")
    }

    service := issues.NewFake()
    handler := api.NewHandler(service)

    mux := http.NewServeMux()
    mux.HandleFunc("/health", healthHandler)
    mux.Handle("POST /repos/{owner}/{repo}/issues", api.RequireAuth(apiToken, http.HandlerFunc(handler.CreateIssue)))
    mux.Handle("GET /repos/{owner}/{repo}/issues", api.RequireAuth(apiToken, http.HandlerFunc(handler.ListIssues)))
    mux.Handle("DELETE /repos/{owner}/{repo}/issues/{number}", api.RequireAuth(apiToken, http.HandlerFunc(handler.CloseIssue)))

    server := &http.Server{
        Addr:    ":8080",
        Handler: mux,
        ReadHeaderTimeout:  5 * time.Second,
        ReadTimeout:      10 * time.Second,
        WriteTimeout:     10 * time.Second,
        IdleTimeout:      60 * time.Second,
    }

    log.Println("Starting server on port :8080")
    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}

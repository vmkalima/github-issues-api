package main

import (
    "log"
    "net/http"

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

    service := issues.NewFake()
    handler := api.NewHandler(service)

    mux := http.NewServeMux()
    mux.HandleFunc("/health", healthHandler)
    mux.HandleFunc("POST /repos/{owner}/{repo}/issues", handler.CreateIssue)

    log.Println("Starting server on port :8080")
    if err := http.ListenAndServe(":8080", mux); err != nil {
        log.Fatal(err)
    }
}

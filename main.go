package main

import (
    "log"
    "net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    if _, err := w.Write([]byte("ok")); err != nil {
        log.Printf("Failed to write response: %v", err)
    }
}

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/health", healthHandler)

    log.Println("Starting server on port :8080")
    if err := http.ListenAndServe(":8080", mux); err != nil {
        log.Fatal(err)
    }
}

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func passthroughHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

// TestRequireAuthMissingHeader tests that the RequireAuth middleware correctly rejects requests that do not include an Authorization header.
func TestRequireAuthMissingHeader(t *testing.T) {
	protected := RequireAuth("valid-token", passthroughHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	protected.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestRequireAuthInvalidToken tests that the RequireAuth middleware correctly rejects requests with an invalid token.
func TestRequireAuthInvalidToken(t *testing.T) {
	protected := RequireAuth("valid-token", passthroughHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	protected.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status code %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// TestRequireAuthValidToken tests that the RequireAuth middleware allows requests with a valid token to pass through to the next handler.
func TestRequireAuthValidToken(t *testing.T) {
	protected := RequireAuth("valid-token", passthroughHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	protected.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}
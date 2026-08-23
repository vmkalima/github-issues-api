package api
import (
	"crypto/subtle"
	"net/http"
)

// RequireAuth wraps an http.Handler, rejecting any request that does not
// carry a valid bearer token matching the configured API token.
//
// The comparison uses a constant-time check (crypto/subtle) rather than a
// plain string comparison, to avoid leaking timing information that could
// help an attacker guess the token one byte at a time.
func RequireAuth(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		const prefix = "Bearer "

		if len(authHeader) <= len(prefix) || authHeader[:len(prefix)] != prefix {
			http.Error(w, "Missing or malformed authorization header", http.StatusUnauthorized)
			return
		}

		requestToken := authHeader[len(prefix):]
		if subtle.ConstantTimeCompare([]byte(requestToken), []byte(token)) != 1 {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
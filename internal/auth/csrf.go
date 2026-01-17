package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

func CSRFMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			next(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			next(w, r)
			return
		}
		next(w, r)
	}
}

func GenerateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}


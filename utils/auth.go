package utils

import (
	"net/http"
	"strings"
)

// Authorized indicates whether an HTTP request is authorized to trigger the
// hook, either through the "Authorization" header or the "authorization" URL
// query parameter.
func Authorized(auth string, req *http.Request) bool {
	if auth == "" {
		return true
	}

	header := req.Header.Get("Authorization")
	if header != "" {
		token := strings.TrimPrefix(header, "Bearer ")
		if token != header && token == auth {
			return true
		}
	}

	query := req.URL.Query()["authorization"]
	for _, value := range query {
		if value == auth {
			return true
		}
	}

	return false
}

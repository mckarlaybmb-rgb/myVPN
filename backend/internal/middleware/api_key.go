package middleware

import "net/http"

func APIKey(expected string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if expected == "" || request.Header.Get("X-API-Key") != expected {
				http.Error(writer, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(writer, request)
		})
	}
}

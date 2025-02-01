package middleware

import (
	"log/slog"
	"net/http"
)

func NewAuth(token string, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Telegram-Bot-Api-Secret-Token") != token {
				logger.Warn("attempt of unauthenticated access", "headers", r.Header)
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

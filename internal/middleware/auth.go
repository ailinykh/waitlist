package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ailinykh/waitlist/internal/clock"
	"github.com/ailinykh/waitlist/pkg/jwt"
)

func JwtAuth(jwtSecret string, contextKey any, clock clock.Clock, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if len(auth) < 7 {
				logger.Error("no auth header passed", slog.String("authorization", auth))
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			logger.Info("got token", slog.String("auth", auth[7:]))
			claims, err := jwt.Decode(auth[7:], jwtSecret)
			if err != nil {
				logger.Error("failed to parse jwt token", slog.Any("error", err))
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			ttl := int64(claims["ttl"].(float64))
			if ttl < clock.Now().Unix() {
				logger.Error("jwt expired", slog.Int64("ttl", ttl))
				http.Redirect(w, r, "/logout", http.StatusFound)
				return
			}

			logger.Info("got valid claims", slog.Any("claims", claims))

			ctx := context.WithValue(r.Context(), contextKey, claims["payload"])
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func HeaderAuth(header, token string, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get(header) != token {
				logger.Warn("attempt of unauthenticated access", slog.Any("headers", r.Header))
				http.Error(w, "Forbidden", http.StatusForbidden)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

func RoleAuth(role string, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			dict, ok := r.Context().Value(User{}).(map[string]interface{})
			if !ok {
				logger.Error("failed to read user dictionary from context", slog.Any("value", r.Context().Value(User{})))
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			user, err := DecodeUser(dict)
			if err != nil {
				logger.Error("failed to decode user from dictionary", slog.Any("value", dict), slog.Any("error", err))
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			if user.Role != role {
				logger.Error("unauthorized access", slog.Any("user", user))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type User struct {
	ID        uint64 `json:"id"`
	UserID    int64  `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	Role      string `json:"role"`
}

func DecodeUser(dict map[string]interface{}) (*User, error) {
	var user = &User{}
	// FIXME: claims type mapping
	if v, ok := dict["id"].(float64); ok {
		user.ID = uint64(v)
	} else {
		return nil, fmt.Errorf("failed to parse ID")
	}
	// FIXME: claims type mapping
	if v, ok := dict["user_id"].(float64); ok {
		user.UserID = int64(v)
	} else {
		return nil, fmt.Errorf("failed to parse UserID")
	}

	if v, ok := dict["first_name"].(string); ok {
		user.FirstName = v
	} else {
		return nil, fmt.Errorf("failed to parse FirstName")
	}

	if v, ok := dict["last_name"].(string); ok {
		user.LastName = v
	} else {
		return nil, fmt.Errorf("failed to parse LastName")
	}

	if v, ok := dict["username"].(string); ok {
		user.Username = v
	} else {
		return nil, fmt.Errorf("failed to parse Username")
	}

	if v, ok := dict["role"].(string); ok {
		user.Role = v
	} else {
		return nil, fmt.Errorf("failed to parse Role")
	}

	return user, nil
}

package app

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"text/template"
	"time"

	"github.com/ailinykh/waitlist/internal/api/telegram"
	"github.com/ailinykh/waitlist/internal/clock"
	"github.com/ailinykh/waitlist/internal/middleware"
	"github.com/ailinykh/waitlist/internal/repository"
	"github.com/ailinykh/waitlist/pkg/jwt"
)

func NewLoginHandlerFunc(bot string, tmpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "login.html", struct {
			BotUsername string
			URL         string
		}{
			BotUsername: bot,
			URL:         fmt.Sprintf("https://%s/api/telegram/callback", r.Host),
		})
	}
}

func NewLogutHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "auth",
			Value:    "",
			MaxAge:   -1,
			Path:     "/",
			HttpOnly: true, // can't be accessed by JavaScript
		})
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func NewCallbackHandlerFunc(config *Config, repo Repo, clock clock.Clock, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("callback",
			slog.String("path", r.URL.Path),
			slog.String("raw", r.URL.RawQuery),
			slog.Any("query", r.URL.Query()),
		)

		values := r.URL.Query()

		if telegram.CalculateHash(values, config.telegramBotToken) != values.Get("hash") {
			logger.Error("❌ checksum mismatch", slog.String("query", r.URL.RawQuery))
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		logger.Info("got new callback", slog.String("auth_date", values.Get("auth_date")))

		userID, err := strconv.ParseInt(values.Get("id"), 10, 64)
		if err != nil {
			logger.Error("failed to parse user_id", slog.Any("error", err))
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		var user repository.User
		user, err = repo.GetUserByUserID(r.Context(), userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				_, err := repo.CreateUser(r.Context(), repository.CreateUserParams{
					UserID:    userID,
					FirstName: values.Get("first_name"),
					LastName:  values.Get("last_name"),
					Username:  values.Get("username"),
					PhotoUrl:  values.Get("photo_url"),
				})
				if err != nil {
					logger.Error("failed to create user", slog.Any("error", err), slog.String("query", r.URL.RawQuery))
				} else {
					user, _ = repo.GetUserByUserID(r.Context(), userID)
				}
			} else {
				logger.Error("failed to find user by id", slog.Any("error", err), slog.Int64("user_id", userID))
			}
		}

		ttl := clock.Now().Add(time.Hour * 24 * 100)

		// logger.Info("ttl", "ttl", ttl.String())
		// tokenString, err := token.SignedString([]byte(config.jwtSecret))

		tokenString, err := jwt.Encode(config.jwtSecret, map[string]interface{}{
			"payload": middleware.User{
				ID:        user.ID,
				UserID:    user.UserID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Username:  user.Username,
				Role:      user.Role,
			},
			"ttl": ttl.Unix(),
		})

		if err != nil {
			logger.Error("failed sign jwt token", slog.Any("error", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		cookie := &http.Cookie{
			Name:     "auth",
			Value:    tokenString,
			Expires:  ttl,
			Path:     "/",
			Secure:   true,
			HttpOnly: true, // can't be accessed by JavaScript
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

package app

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func NewAPIHandlerFunc(logger *slog.Logger, repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		waitlist, err := repo.GetAllEntries(r.Context())
		if err != nil {
			logger.Error("failed to get waitlist", slog.Any("error", err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		logger.Info("get all", slog.Int("count", len(waitlist)))

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(waitlist)

		if err != nil {
			logger.Error("failed to encode waitlist", slog.Any("error", err))
		}
	}
}

package app

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func NewAPIHandlerFunc(logger *slog.Logger, repo Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		waitlist, err := repo.GetAll(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("failed to get waitlist", "error", err)
			return
		}

		logger.Info("Get all", slog.Int("count", len(waitlist)))

		data, err := json.Marshal(waitlist)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("failed marshal waitlist", "error", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

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
			logger.Error("failed to get waitlist", slog.Any("error", err))
			return
		}

		logger.Info("get all", slog.Int("count", len(waitlist)))

		data, err := json.Marshal(waitlist)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("failed marshal waitlist", slog.Any("error", err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/pkg/response"
)

type SyncHandler struct {
	syncService service.SyncServiceInterface
	log         zerolog.Logger
}

func NewSyncHandler(syncSvc service.SyncServiceInterface, log zerolog.Logger) *SyncHandler {
	return &SyncHandler{
		syncService: syncSvc,
		log:         log,
	}
}

func (h *SyncHandler) Pull(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	if storeID == "" {
		response.Error(w, http.StatusBadRequest, "Missing store ID")
		return
	}

	lastSyncedAtStr := r.URL.Query().Get("last_synced_at")
	var since time.Time
	if lastSyncedAtStr != "" {
		parsed, err := time.Parse(time.RFC3339, lastSyncedAtStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid last_synced_at format. Use RFC3339.")
			return
		}
		since = parsed
	}

	res, err := h.syncService.Pull(r.Context(), storeID, since)
	if err != nil {
		h.log.Error().Err(err).Str("store_id", storeID).Msg("Sync pull failed")
		response.Error(w, http.StatusInternalServerError, "Failed to pull sync data")
		return
	}

	response.Success(w, res)
}

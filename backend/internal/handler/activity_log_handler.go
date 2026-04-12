package handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/moedahpos/backend/internal/dto"
	"github.com/moedahpos/backend/internal/service"
	"github.com/moedahpos/backend/pkg/response"
)

type ActivityLogHandler struct {
	activitySvc *service.ActivityLogService
	log         zerolog.Logger
}

func NewActivityLogHandler(activitySvc *service.ActivityLogService, log zerolog.Logger) *ActivityLogHandler {
	return &ActivityLogHandler{activitySvc: activitySvc, log: log}
}

func (h *ActivityLogHandler) List(w http.ResponseWriter, r *http.Request) {
	storeID := chi.URLParam(r, "storeId")
	filter := dto.ActivityLogFilter{StoreID: storeID}

	// Manual decode query params
	q := r.URL.Query()
	filter.Page, _ = strconv.Atoi(q.Get("page"))
	filter.PerPage, _ = strconv.Atoi(q.Get("per_page"))
	filter.UserID = q.Get("user_id")
	filter.Module = q.Get("module")
	filter.ActionType = q.Get("action_type")
	filter.StartDate = q.Get("start_date")
	filter.EndDate = q.Get("end_date")

	logs, meta, err := h.activitySvc.GetActivityLogs(r.Context(), storeID, filter)
	if err != nil {
		h.log.Error().Err(err).Msg("failed to get activity logs")
		response.InternalError(w)
		return
	}

	response.Success(w, dto.ListResponse{
		Data: logs,
		Meta: meta,
	})
}

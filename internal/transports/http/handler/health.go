package handler

import (
	"authentication_service/internal/logger"
	"authentication_service/internal/service"
	"encoding/json"
	"net/http"
	"time"
)

type Opts struct {
	HealthService service.HealthService
	Logger        logger.Logger
}

type HealthHandler struct {
	healthService service.HealthService
	logger        logger.Logger
}

func NewHealthHandler(opts *Opts) *HealthHandler {
	return &HealthHandler{
		healthService: opts.HealthService,
		logger:        opts.Logger,
	}
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "Auth-Service")
	w.Header().Set("Date", time.Now().Format(time.RFC1123))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeJSON(w, map[string]string{"status": "ok"}, h.logger)
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	status := h.healthService.Check(r.Context())
	w.Header().Set("Server", "Auth-Service")
	w.Header().Set("Date", time.Now().Format(time.RFC1123))
	w.Header().Set("Content-Type", "application/json")

	if status.Status == "ready" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	writeJSON(w, status, h.logger)
}

func writeJSON(w http.ResponseWriter, data any, logger logger.Logger) {
	w.Header().Set("Server", "Auth-Service")
	w.Header().Set("Date", time.Now().Format(time.RFC1123))
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("failed to write json: " + err.Error())
	}
}

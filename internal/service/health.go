package service

import (
	"authentication_service/internal/cache"
	"authentication_service/internal/database"
	"context"
	"sync/atomic"

	"gorm.io/gorm"
)

type HealthService interface {
	Check(ctx context.Context) HealthStatus
	SetReady(ready bool)
}

type HealthStatus struct {
	Status  string            `json:"status"`
	Details map[string]string `json:"details,omitempty"`
}

type healthService struct {
	database *gorm.DB
	cache    cache.CacheService
	ready    atomic.Bool
}

func (hs *healthService) SetReady(ready bool) {
	hs.ready.Store(ready)
}

func (hs *healthService) Check(ctx context.Context) HealthStatus {
	if !hs.ready.Load() {
		return HealthStatus{
			Status:  "unready",
			Details: map[string]string{"service": "shutting down"},
		}
	}

	status := HealthStatus{Status: "ready", Details: map[string]string{}}

	if err := hs.database.Exec("SELECT 1").Error; err != nil {
		status.Status = "unready"
		status.Details["database"] = err.Error()
	} else {
		status.Details["database"] = "ok"
	}

	if err := hs.cache.Ping(ctx); err != nil {
		status.Status = "unready"
		status.Details["cache"] = err.Error()
	} else {
		status.Details["cache"] = "ok"
	}

	return status
}

type HealthServiceOpts struct {
	Database database.DatabaseService
	Cache    cache.CacheService
}

func NewHealthService(opts *HealthServiceOpts) HealthService {
	h := &healthService{
		database: opts.Database.DB(),
		cache:    opts.Cache,
	}
	h.ready.Store(true)
	return h
}

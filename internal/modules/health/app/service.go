package app

import (
	"context"
	"time"
)

type DependencyChecker interface {
	Check(context.Context) error
}

type ComponentStatus struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	LatencyMs int64  `json:"latency_ms"`
}

type HealthStatus struct {
	Healthy    bool                       `json:"healthy"`
	Components map[string]ComponentStatus `json:"components"`
}

type Service struct {
	db    DependencyChecker
	redis DependencyChecker
	s3    DependencyChecker
}

func NewService(db, redis, s3 DependencyChecker) *Service {
	return &Service{db: db, redis: redis, s3: s3}
}

func (s *Service) Check(ctx context.Context) HealthStatus {
	if ctx == nil {
		ctx = context.Background()
	}

	status := HealthStatus{
		Healthy:    true,
		Components: map[string]ComponentStatus{},
	}

	dbStatus := s.check(ctx, s.db, "database not initialized", false)
	status.Components["database"] = dbStatus
	if dbStatus.Status != "ok" {
		status.Healthy = false
	}

	redisStatus := s.check(ctx, s.redis, "redis not initialized", false)
	status.Components["redis"] = redisStatus
	if redisStatus.Status != "ok" {
		status.Healthy = false
	}

	status.Components["s3"] = s.check(ctx, s.s3, "client not initialized", true)
	return status
}

func (s *Service) check(ctx context.Context, checker DependencyChecker, nilMessage string, informational bool) ComponentStatus {
	start := time.Now()
	if checker == nil {
		if informational {
			return ComponentStatus{Status: "degraded", Message: nilMessage, LatencyMs: time.Since(start).Milliseconds()}
		}
		return ComponentStatus{Status: "error", Message: nilMessage, LatencyMs: time.Since(start).Milliseconds()}
	}

	if err := checker.Check(ctx); err != nil {
		if informational {
			return ComponentStatus{Status: "degraded", Message: err.Error(), LatencyMs: time.Since(start).Milliseconds()}
		}
		return ComponentStatus{Status: "error", Message: err.Error(), LatencyMs: time.Since(start).Milliseconds()}
	}

	return ComponentStatus{Status: "ok", Message: "reachable", LatencyMs: time.Since(start).Milliseconds()}
}

package app

import (
	"context"
	"errors"
	"testing"
)

type stubChecker struct {
	err error
}

func (s stubChecker) Check(_ context.Context) error { return s.err }

func TestHealthServiceMarksDatabaseAndRedisAsCoreDependencies(t *testing.T) {
	service := NewService(stubChecker{err: errors.New("database not initialized")}, stubChecker{err: errors.New("redis not initialized")}, stubChecker{err: errors.New("client not initialized")})

	status := service.Check(context.Background())
	if status.Healthy {
		t.Fatalf("expected unhealthy status when database and redis are unavailable")
	}
	if status.Components["database"].Status != "error" {
		t.Fatalf("expected database error status, got %+v", status.Components["database"])
	}
	if status.Components["redis"].Status != "error" {
		t.Fatalf("expected redis error status, got %+v", status.Components["redis"])
	}
	if status.Components["s3"].Status != "degraded" {
		t.Fatalf("expected s3 degraded status, got %+v", status.Components["s3"])
	}
}

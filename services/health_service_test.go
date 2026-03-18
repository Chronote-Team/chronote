package services

import (
	"context"
	"testing"

	"chronote/config"
	"chronote/global"
)

func TestHealthServiceCheckWithoutDependencies(t *testing.T) {
	originalDB := global.Db
	originalRedis := global.RedisClient
	originalS3 := config.S3Client
	originalConfig := config.AppConfig
	t.Cleanup(func() {
		global.Db = originalDB
		global.RedisClient = originalRedis
		config.S3Client = originalS3
		config.AppConfig = originalConfig
	})

	global.Db = nil
	global.RedisClient = nil
	config.S3Client = nil
	config.AppConfig = nil

	service := HealthService{}
	status := service.Check(context.Background())

	if status.Healthy {
		t.Fatalf("expected unhealthy status when db and redis are unavailable")
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

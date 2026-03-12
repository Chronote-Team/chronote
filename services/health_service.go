package services

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"

	"chronote/config"
	"chronote/global"
)

type ComponentStatus struct {
	Status    string `json:"status"` // "ok" | "degraded" | "error"
	Message   string `json:"message"`
	LatencyMs int64  `json:"latency_ms"`
}

type HealthStatus struct {
	Healthy    bool                       `json:"healthy"`
	Components map[string]ComponentStatus `json:"components"`
}

type HealthService struct{}

const healthCheckTimeout = 2 * time.Second

func (s *HealthService) Check(ctx context.Context) HealthStatus {
	if ctx == nil {
		ctx = context.Background()
	}

	result := HealthStatus{
		Healthy:    true,
		Components: make(map[string]ComponentStatus),
	}

	dbStatus := s.checkDB(ctx)
	result.Components["database"] = dbStatus
	if dbStatus.Status != "ok" {
		result.Healthy = false
	}

	redisStatus := s.checkRedis(ctx)
	result.Components["redis"] = redisStatus
	if redisStatus.Status != "ok" {
		result.Healthy = false
	}

	// S3 is informational and does not affect overall health.
	result.Components["s3"] = s.checkS3(ctx)

	return result
}

func (s *HealthService) checkDB(ctx context.Context) ComponentStatus {
	start := time.Now()
	tctx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	if global.Db == nil {
		return ComponentStatus{Status: "error", Message: "database not initialized", LatencyMs: elapsed(start)}
	}

	sqlDB, err := global.Db.DB()
	if err != nil {
		return ComponentStatus{Status: "error", Message: err.Error(), LatencyMs: elapsed(start)}
	}
	if err := sqlDB.PingContext(tctx); err != nil {
		return ComponentStatus{Status: "error", Message: err.Error(), LatencyMs: elapsed(start)}
	}
	return ComponentStatus{Status: "ok", Message: "reachable", LatencyMs: elapsed(start)}
}

func (s *HealthService) checkRedis(ctx context.Context) ComponentStatus {
	start := time.Now()
	tctx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	if global.RedisClient == nil {
		return ComponentStatus{Status: "error", Message: "redis not initialized", LatencyMs: elapsed(start)}
	}

	if err := global.RedisClient.Ping(tctx).Err(); err != nil {
		return ComponentStatus{Status: "error", Message: err.Error(), LatencyMs: elapsed(start)}
	}
	return ComponentStatus{Status: "ok", Message: "reachable", LatencyMs: elapsed(start)}
}

func (s *HealthService) checkS3(ctx context.Context) ComponentStatus {
	start := time.Now()
	tctx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	if config.S3Client == nil {
		return ComponentStatus{Status: "degraded", Message: "client not initialized", LatencyMs: elapsed(start)}
	}

	bucketName := strings.TrimSpace(config.AppConfig.S3.BucketName)
	if bucketName == "" {
		return ComponentStatus{Status: "degraded", Message: "bucket not configured", LatencyMs: elapsed(start)}
	}

	_, err := config.S3Client.HeadBucket(tctx, &s3.HeadBucketInput{Bucket: &bucketName})
	if err != nil {
		return ComponentStatus{Status: "degraded", Message: err.Error(), LatencyMs: elapsed(start)}
	}

	return ComponentStatus{Status: "ok", Message: "reachable", LatencyMs: elapsed(start)}
}

func elapsed(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}

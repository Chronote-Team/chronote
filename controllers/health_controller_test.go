package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chronote/config"
	"chronote/global"

	"github.com/gin-gonic/gin"
)

func TestHealthCheckReturnsServiceUnavailableWhenCoreDepsMissing(t *testing.T) {
	restoreHealthGlobals(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = req

	HealthCheck(ctx)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body["message"] != "Service Unavailable" {
		t.Fatalf("expected service unavailable message, got %#v", body["message"])
	}
}

func TestHealthDetailsReturnsComponentStatusPayload(t *testing.T) {
	restoreHealthGlobals(t)

	req := httptest.NewRequest(http.MethodGet, "/health/details", nil)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = req

	HealthDetails(ctx)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
	}

	var body struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Healthy    bool `json:"healthy"`
			Components map[string]struct {
				Status string `json:"status"`
			} `json:"components"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if body.Code != http.StatusServiceUnavailable || body.Message != "Service Unavailable" {
		t.Fatalf("unexpected response envelope: %+v", body)
	}
	if body.Data.Healthy {
		t.Fatalf("expected unhealthy detail payload")
	}
	if body.Data.Components["database"].Status != "error" {
		t.Fatalf("expected database error status, got %+v", body.Data.Components["database"])
	}
	if body.Data.Components["redis"].Status != "error" {
		t.Fatalf("expected redis error status, got %+v", body.Data.Components["redis"])
	}
	if body.Data.Components["s3"].Status != "degraded" {
		t.Fatalf("expected s3 degraded status, got %+v", body.Data.Components["s3"])
	}
}

func restoreHealthGlobals(t *testing.T) {
	t.Helper()

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

	gin.SetMode(gin.TestMode)
	global.Db = nil
	global.RedisClient = nil
	config.S3Client = nil
	config.AppConfig = nil
}

package config

import "testing"

func TestLoadReadsAIEndpointFromEnv(t *testing.T) {
	t.Setenv("CONFIG_PATH", t.TempDir())
	t.Setenv("AI_ENDPOINT", "https://proxy.example.com/v1/responses")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AI.Endpoint != "https://proxy.example.com/v1/responses" {
		t.Fatalf("AI.Endpoint = %q, want %q", cfg.AI.Endpoint, "https://proxy.example.com/v1/responses")
	}
}

func TestLoadReadsAIWorkerSettingsFromEnv(t *testing.T) {
	t.Setenv("CONFIG_PATH", t.TempDir())
	t.Setenv("AI_WORKER_ID", "worker-2")
	t.Setenv("AI_WORKER_IDLE_SLEEP", "25ms")
	t.Setenv("AI_WORKER_ERROR_SLEEP", "50ms")
	t.Setenv("AI_WORKER_RUN_ONCE", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.AIWorker.ID != "worker-2" {
		t.Fatalf("AIWorker.ID = %q, want worker-2", cfg.AIWorker.ID)
	}
	if cfg.AIWorker.IdleSleep != "25ms" {
		t.Fatalf("AIWorker.IdleSleep = %q, want 25ms", cfg.AIWorker.IdleSleep)
	}
	if cfg.AIWorker.ErrorSleep != "50ms" {
		t.Fatalf("AIWorker.ErrorSleep = %q, want 50ms", cfg.AIWorker.ErrorSleep)
	}
	if !cfg.AIWorker.RunOnce {
		t.Fatal("AIWorker.RunOnce = false, want true")
	}
}

func TestLoadReadsS3PublicBaseURLFromEnv(t *testing.T) {
	t.Setenv("CONFIG_PATH", t.TempDir())
	t.Setenv("S3_PUBLIC_BASE_URL", "https://media.example.com/bucket")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.S3.PublicBaseURL != "https://media.example.com/bucket" {
		t.Fatalf("S3.PublicBaseURL = %q, want %q", cfg.S3.PublicBaseURL, "https://media.example.com/bucket")
	}
}

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

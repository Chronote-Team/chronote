package postcardai_test

import (
	"encoding/json"
	"testing"
	"time"

	"chronote-refactor/internal/modules/postcardai/domain"
	aiinfra "chronote-refactor/internal/modules/postcardai/infra/ai"
)

func TestAnalysisStatusTransitions(t *testing.T) {
	tests := []struct {
		name string
		from domain.AnalysisStatus
		to   domain.AnalysisStatus
		want bool
	}{
		{name: "pending can process", from: domain.StatusPending, to: domain.StatusProcessing, want: true},
		{name: "processing can succeed", from: domain.StatusProcessing, to: domain.StatusSucceeded, want: true},
		{name: "processing can retry", from: domain.StatusProcessing, to: domain.StatusPending, want: true},
		{name: "succeeded cannot process again", from: domain.StatusSucceeded, to: domain.StatusProcessing, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domain.CanTransition(tt.from, tt.to); got != tt.want {
				t.Fatalf("CanTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestVersionKeyEquality(t *testing.T) {
	key := domain.VersionKey{
		ResourceID:      7,
		ResourceVersion: "2026-05-13T00:00:00Z",
		PromptVersion:   "image_understanding_v1",
		SchemaVersion:   "image_understanding_v1",
		ModelVersion:    "gpt-4.1-mini",
	}
	if !key.Equal(key) {
		t.Fatal("expected identical version keys to match")
	}

	changed := key
	changed.SchemaVersion = "image_understanding_v2"
	if key.Equal(changed) {
		t.Fatal("expected schema version mismatch to make version keys unequal")
	}
}

func TestSchemaValidationRejectsMissingRequiredFields(t *testing.T) {
	if err := aiinfra.ValidateImageUnderstanding(json.RawMessage(`{"caption":"food"}`)); err == nil {
		t.Fatal("expected missing image_type to fail validation")
	}
	if err := aiinfra.ValidatePostcardUnderstanding(json.RawMessage(`{"summary":"trip"}`)); err == nil {
		t.Fatal("expected missing suggested_title to fail validation")
	}
}

func TestAnalysisJobRejectsStaleCompletion(t *testing.T) {
	job := domain.AnalysisJob{
		ID:              1,
		PostcardID:      9,
		PostcardVersion: "old",
		Status:          domain.StatusProcessing,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if job.IsCurrentFor("new") {
		t.Fatal("expected old postcard version to be stale")
	}
}

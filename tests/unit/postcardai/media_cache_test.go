package postcardai_test

import (
	"context"
	"encoding/json"
	"testing"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
)

func TestWorkerReusesMediaAnalysisForMatchingVersionKey(t *testing.T) {
	jobs := newFakeJobRepository()
	results := newFakeResultRepository()
	results.media[11] = domain.MediaAnalysis{
		MediaID:       11,
		MediaVersion:  "mv1",
		PromptVersion: postcardaiapp.DefaultImagePromptVersion,
		SchemaVersion: postcardaiapp.DefaultImageSchemaVersion,
		ModelVersion:  "test-model",
		Status:        domain.StatusSucceeded,
		Result:        json.RawMessage(`{"image_type":"food","caption":"ramen","confidence":0.9}`),
		Confidence:    0.9,
	}
	provider := &countingAIClient{}
	service := postcardaiapp.NewService(postcardaiapp.Dependencies{
		Jobs:      jobs,
		Results:   results,
		Postcards: fakePostcardSource{version: "pv2"},
		Media:     fakeMediaSource{items: []postcardaiapp.MediaSnapshot{{ID: 11, Version: "mv1", Type: "image", StorageKey: "media/11.jpg"}}},
		Storage:   fakeStorage{},
		AI:        provider,
		Enabled:   true,
		Model:     "test-model",
	})
	if _, err := service.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{PostcardID: 3, Reason: postcardaiapp.EnqueueReasonUpdate}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if _, err := service.RunNextAnalysisJob(context.Background(), "worker-1"); err != nil {
		t.Fatalf("run: %v", err)
	}
	if provider.imageCalls != 0 {
		t.Fatalf("expected media analysis reuse, got %d image calls", provider.imageCalls)
	}
}

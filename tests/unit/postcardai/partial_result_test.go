package postcardai_test

import (
	"context"
	"encoding/json"
	"testing"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
)

func TestWorkerStoresPartialPostcardUnderstandingWhenOneImageFails(t *testing.T) {
	provider := &partialAIClient{}
	results := newFakeResultRepository()
	service := postcardaiapp.NewService(postcardaiapp.Dependencies{
		Jobs:      newFakeJobRepository(),
		Results:   results,
		Postcards: fakePostcardSource{version: "pv1", content: json.RawMessage(`{"text":"trip"}`)},
		Media: fakeMediaSource{items: []postcardaiapp.MediaSnapshot{
			{ID: 11, Version: "mv1", Type: "image", StorageKey: "media/11.jpg"},
			{ID: 12, Version: "mv1", Type: "image", StorageKey: "media/12.jpg"},
		}},
		Storage: fakeStorage{},
		AI:      provider,
		Enabled: true,
	})
	if _, err := service.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{PostcardID: 3, Reason: postcardaiapp.EnqueueReasonCreate}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	outcome, err := service.RunNextAnalysisJob(context.Background(), "worker-1")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if outcome.Status != domain.StatusSucceeded {
		t.Fatalf("expected partial final success, got %q", outcome.Status)
	}
	if len(results.postcards) != 1 || results.postcards[0].Uncertainty == "" {
		t.Fatalf("expected postcard analysis with uncertainty, got %#v", results.postcards)
	}
}

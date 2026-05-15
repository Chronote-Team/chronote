package postcardai_test

import (
	"context"
	"encoding/json"
	"testing"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
)

func TestWorkerPerformsOneSchemaRepairRetry(t *testing.T) {
	provider := &repairAIClient{
		firstImage:    json.RawMessage(`{"caption":"bad"}`),
		repairedImage: json.RawMessage(`{"image_type":"food","caption":"ramen","confidence":0.8}`),
		postcard:      json.RawMessage(`{"summary":"lunch","suggested_title":"Lunch","confidence":0.8}`),
	}
	jobs := newFakeJobRepository()
	service := postcardaiapp.NewService(postcardaiapp.Dependencies{
		Jobs:      jobs,
		Results:   newFakeResultRepository(),
		Postcards: fakePostcardSource{version: "pv1"},
		Media:     fakeMediaSource{items: []postcardaiapp.MediaSnapshot{{ID: 11, Version: "mv1", Type: "image", StorageKey: "media/11.jpg"}}},
		Storage:   fakeStorage{},
		AI:        provider,
		Enabled:   true,
	})
	if _, err := service.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{PostcardID: 3, Reason: postcardaiapp.EnqueueReasonCreate}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	outcome, err := service.RunNextAnalysisJob(context.Background(), "worker-1")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if outcome.Status != domain.StatusSucceeded {
		t.Fatalf("expected succeeded after repair, got %q", outcome.Status)
	}
	if provider.repairImageCalls != 1 {
		t.Fatalf("expected one repair call, got %d", provider.repairImageCalls)
	}
}

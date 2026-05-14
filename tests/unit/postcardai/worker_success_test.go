package postcardai_test

import (
	"context"
	"encoding/json"
	"testing"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
)

func TestWorkerStoresImageAndPostcardAnalysis(t *testing.T) {
	jobs := newFakeJobRepository()
	results := newFakeResultRepository()
	postcards := fakePostcardSource{version: "pv1", content: json.RawMessage(`{"blocks":[{"type":"text","value":"lunch"}]}`)}
	medias := fakeMediaSource{items: []postcardaiapp.MediaSnapshot{{ID: 11, Version: "mv1", Type: "image", StorageKey: "media/11.jpg"}}}
	storage := fakeStorage{}
	provider := fakeAIClient{}

	service := postcardaiapp.NewService(postcardaiapp.Dependencies{
		Jobs:      jobs,
		Results:   results,
		Postcards: postcards,
		Media:     medias,
		Storage:   storage,
		AI:        provider,
		Enabled:   true,
	})
	if _, err := service.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{PostcardID: 3, Reason: postcardaiapp.EnqueueReasonCreate}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	outcome, err := service.RunNextAnalysisJob(context.Background(), "worker-1")
	if err != nil {
		t.Fatalf("RunNextAnalysisJob returned error: %v", err)
	}
	if outcome.Status != domain.StatusSucceeded {
		t.Fatalf("expected succeeded outcome, got %q", outcome.Status)
	}
	if len(results.media) != 1 {
		t.Fatalf("expected one media analysis, got %d", len(results.media))
	}
	if len(results.postcards) != 1 {
		t.Fatalf("expected one postcard analysis, got %d", len(results.postcards))
	}
	if jobs.jobs[0].Status != domain.StatusSucceeded {
		t.Fatalf("expected job succeeded, got %q", jobs.jobs[0].Status)
	}
}

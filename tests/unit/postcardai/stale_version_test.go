package postcardai_test

import (
	"context"
	"testing"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
)

func TestWorkerMarksJobStaleWhenPostcardVersionChangesBeforeStore(t *testing.T) {
	source := &changingPostcardSource{versions: []string{"old", "new"}}
	jobs := newFakeJobRepository()
	service := postcardaiapp.NewService(postcardaiapp.Dependencies{
		Jobs:      jobs,
		Results:   newFakeResultRepository(),
		Postcards: source,
		Media:     fakeMediaSource{},
		Storage:   fakeStorage{},
		AI:        fakeAIClient{},
		Enabled:   true,
	})
	if _, err := service.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{PostcardID: 3, Reason: postcardaiapp.EnqueueReasonCreate}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	outcome, err := service.RunNextAnalysisJob(context.Background(), "worker-1")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if outcome.Status != domain.StatusStale {
		t.Fatalf("expected stale, got %q", outcome.Status)
	}
}

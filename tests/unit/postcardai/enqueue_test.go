package postcardai_test

import (
	"context"
	"testing"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
)

func TestEnqueueCreatesPendingJobWhenEligible(t *testing.T) {
	repo := newFakeJobRepository()
	service := postcardaiapp.NewService(postcardaiapp.Dependencies{
		Jobs:      repo,
		Postcards: fakePostcardSource{version: "v1"},
		Enabled:   true,
	})

	result, err := service.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{
		PostcardID: 3,
		Reason:     postcardaiapp.EnqueueReasonCreate,
	})
	if err != nil {
		t.Fatalf("EnqueuePostcardAnalysis returned error: %v", err)
	}
	if result.Noop {
		t.Fatal("expected active enqueue, got noop")
	}
	if result.Job.Status != domain.StatusPending {
		t.Fatalf("expected pending job, got %q", result.Job.Status)
	}
	if result.Job.PostcardVersion != "v1" {
		t.Fatalf("expected postcard version v1, got %q", result.Job.PostcardVersion)
	}
}

func TestEnqueueReturnsExistingActiveJobForSamePostcardVersion(t *testing.T) {
	repo := newFakeJobRepository()
	service := postcardaiapp.NewService(postcardaiapp.Dependencies{
		Jobs:      repo,
		Postcards: fakePostcardSource{version: "v1"},
		Enabled:   true,
	})

	first, err := service.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{PostcardID: 3, Reason: postcardaiapp.EnqueueReasonCreate})
	if err != nil {
		t.Fatalf("first enqueue: %v", err)
	}
	second, err := service.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{PostcardID: 3, Reason: postcardaiapp.EnqueueReasonUpdate})
	if err != nil {
		t.Fatalf("second enqueue: %v", err)
	}
	if !second.Existing {
		t.Fatal("expected second enqueue to return existing job")
	}
	if first.Job.ID != second.Job.ID {
		t.Fatalf("expected same job id, got %d and %d", first.Job.ID, second.Job.ID)
	}
}

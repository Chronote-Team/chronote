package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
)

func TestPostcardAIWorkerInMemoryRecoveryFlow(t *testing.T) {
	jobs := newIntegrationJobRepo()
	results := newIntegrationResultRepo()
	service := postcardaiapp.NewService(postcardaiapp.Dependencies{
		Jobs:      jobs,
		Results:   results,
		Postcards: integrationPostcardSource{},
		Media:     integrationMediaSource{},
		Storage:   postcardaiapp.NoopStorage{},
		AI:        postcardaiapp.NoopAIClient{},
		Enabled:   true,
	})
	if _, err := service.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{PostcardID: 42, Reason: postcardaiapp.EnqueueReasonBackfill}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	outcome, err := service.RunNextAnalysisJob(context.Background(), "integration-worker")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if outcome.Status != domain.StatusSucceeded {
		t.Fatalf("expected succeeded, got %q", outcome.Status)
	}
	if len(results.postcards) != 1 || len(results.media) != 1 {
		t.Fatalf("expected stored postcard/media analyses, got postcards=%d media=%d", len(results.postcards), len(results.media))
	}
}

type integrationJobRepo struct {
	nextID uint
	jobs   []domain.AnalysisJob
}

func newIntegrationJobRepo() *integrationJobRepo {
	return &integrationJobRepo{nextID: 1}
}

func (r *integrationJobRepo) Enqueue(ctx context.Context, job domain.AnalysisJob) (*domain.AnalysisJob, bool, error) {
	job.ID = r.nextID
	r.nextID++
	r.jobs = append(r.jobs, job)
	return &r.jobs[len(r.jobs)-1], false, nil
}

func (r *integrationJobRepo) ClaimNext(ctx context.Context, workerID string, now time.Time) (*domain.AnalysisJob, error) {
	for i := range r.jobs {
		if r.jobs[i].Status == domain.StatusPending {
			r.jobs[i].Status = domain.StatusProcessing
			return &r.jobs[i], nil
		}
	}
	return nil, nil
}

func (r *integrationJobRepo) UpdateStatus(ctx context.Context, id uint, status domain.AnalysisStatus, errorCode domain.ProviderErrorCode) error {
	for i := range r.jobs {
		if r.jobs[i].ID == id {
			r.jobs[i].Status = status
			r.jobs[i].LastErrorCode = errorCode
		}
	}
	return nil
}

func (r *integrationJobRepo) ScheduleRetry(ctx context.Context, id uint, nextRunAt time.Time, errorCode domain.ProviderErrorCode) error {
	return r.UpdateStatus(ctx, id, domain.StatusPending, errorCode)
}

func (r *integrationJobRepo) FindByID(ctx context.Context, id uint) (*domain.AnalysisJob, error) {
	for i := range r.jobs {
		if r.jobs[i].ID == id {
			return &r.jobs[i], nil
		}
	}
	return nil, nil
}

type integrationResultRepo struct {
	media     []domain.MediaAnalysis
	postcards []domain.PostcardAnalysis
}

func newIntegrationResultRepo() *integrationResultRepo {
	return &integrationResultRepo{}
}

func (r *integrationResultRepo) FindReusableMediaAnalysis(ctx context.Context, key domain.VersionKey) (*domain.MediaAnalysis, error) {
	return nil, nil
}

func (r *integrationResultRepo) StoreMediaAnalysis(ctx context.Context, analysis domain.MediaAnalysis) error {
	r.media = append(r.media, analysis)
	return nil
}

func (r *integrationResultRepo) StorePostcardAnalysis(ctx context.Context, analysis domain.PostcardAnalysis) error {
	r.postcards = append(r.postcards, analysis)
	return nil
}

type integrationPostcardSource struct{}

func (integrationPostcardSource) GetPostcardForAnalysis(ctx context.Context, postcardID uint) (*postcardaiapp.PostcardSnapshot, error) {
	return &postcardaiapp.PostcardSnapshot{ID: postcardID, Version: "pv1", Content: json.RawMessage(`{"text":"lunch"}`), UpdatedAt: time.Now()}, nil
}

type integrationMediaSource struct{}

func (integrationMediaSource) ListMediaForPostcard(ctx context.Context, postcardID uint) ([]postcardaiapp.MediaSnapshot, error) {
	return []postcardaiapp.MediaSnapshot{{ID: 99, Version: "mv1", Type: "image", StorageKey: "media/99.jpg"}}, nil
}

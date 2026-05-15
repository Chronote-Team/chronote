package integration

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
	appplatform "chronote-refactor/internal/platform/app"
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

func TestAnalysisWorkerRuntimeConsumesTextOnlyJob(t *testing.T) {
	jobs := newIntegrationJobRepo()
	results := newIntegrationResultRepo()
	service := postcardaiapp.NewService(postcardaiapp.Dependencies{
		Jobs:      jobs,
		Results:   results,
		Postcards: integrationPostcardSource{},
		Media:     integrationMediaSource{items: []postcardaiapp.MediaSnapshot{}},
		Storage:   postcardaiapp.NoopStorage{},
		AI:        postcardaiapp.NoopAIClient{},
		Enabled:   true,
	})
	if _, err := service.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{PostcardID: 10, Reason: postcardaiapp.EnqueueReasonBackfill}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if err := appplatform.RunAnalysisWorker(context.Background(), service, appplatform.WorkerOptions{RunOnce: true}); err != nil {
		t.Fatalf("run worker: %v", err)
	}
	job, _ := jobs.FindByID(context.Background(), 1)
	if job == nil || job.Status != domain.StatusSucceeded {
		t.Fatalf("expected succeeded job, got %#v", job)
	}
	if len(results.postcards) != 1 {
		t.Fatalf("expected stored postcard analysis, got %d", len(results.postcards))
	}
	if len(results.media) != 0 {
		t.Fatalf("expected no media analysis for text-only postcard, got %d", len(results.media))
	}
}

func TestAnalysisWorkerRuntimeLeavesJobsPendingWhenNotRunning(t *testing.T) {
	jobs := newIntegrationJobRepo()
	service := postcardaiapp.NewService(postcardaiapp.Dependencies{
		Jobs:      jobs,
		Postcards: integrationPostcardSource{},
		Enabled:   true,
	})
	if _, err := service.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{PostcardID: 11, Reason: postcardaiapp.EnqueueReasonBackfill}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	job, _ := jobs.FindByID(context.Background(), 1)
	if job == nil || job.Status != domain.StatusPending {
		t.Fatalf("expected pending job without worker, got %#v", job)
	}
}

func TestAnalysisWorkerRuntimeMarksStaleJob(t *testing.T) {
	jobs := newIntegrationJobRepo()
	results := newIntegrationResultRepo()
	source := &changingIntegrationPostcardSource{versions: []string{"v1", "v2"}}
	service := postcardaiapp.NewService(postcardaiapp.Dependencies{
		Jobs:      jobs,
		Results:   results,
		Postcards: source,
		Media:     integrationMediaSource{},
		Storage:   postcardaiapp.NoopStorage{},
		AI:        postcardaiapp.NoopAIClient{},
		Enabled:   true,
	})
	if _, err := service.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{PostcardID: 12, Reason: postcardaiapp.EnqueueReasonBackfill}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if err := appplatform.RunAnalysisWorker(context.Background(), service, appplatform.WorkerOptions{RunOnce: true}); err != nil {
		t.Fatalf("run worker: %v", err)
	}
	job, _ := jobs.FindByID(context.Background(), 1)
	if job == nil || job.Status != domain.StatusStale {
		t.Fatalf("expected stale job, got %#v", job)
	}
	if len(results.postcards) != 0 {
		t.Fatalf("expected no stale postcard analysis writes, got %d", len(results.postcards))
	}
}

func TestAnalysisWorkerRuntimeSchedulesRetryOnProviderUnavailable(t *testing.T) {
	jobs := newIntegrationJobRepo()
	service := postcardaiapp.NewService(postcardaiapp.Dependencies{
		Jobs:      jobs,
		Results:   newIntegrationResultRepo(),
		Postcards: integrationPostcardSource{},
		Media:     integrationMediaSource{},
		Storage:   postcardaiapp.NoopStorage{},
		AI:        unavailableIntegrationAI{},
		Enabled:   true,
	})
	if _, err := service.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{PostcardID: 13, Reason: postcardaiapp.EnqueueReasonBackfill}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if err := appplatform.RunAnalysisWorker(context.Background(), service, appplatform.WorkerOptions{RunOnce: true}); err != nil {
		t.Fatalf("run worker: %v", err)
	}
	job, _ := jobs.FindByID(context.Background(), 1)
	if job == nil || job.Status != domain.StatusPending || job.LastErrorCode != domain.ErrorProviderUnavailable || job.NextRunAt == nil {
		t.Fatalf("expected retryable pending job, got %#v", job)
	}
}

func TestAnalysisWorkerRuntimeReusesMediaAnalysis(t *testing.T) {
	jobs := newIntegrationJobRepo()
	results := newIntegrationResultRepo()
	ai := &countingIntegrationAI{}
	media := integrationMediaSource{items: []postcardaiapp.MediaSnapshot{{ID: 99, Version: "mv1", Type: "image", StorageKey: "media/99.jpg"}}}
	service := postcardaiapp.NewService(postcardaiapp.Dependencies{
		Jobs:      jobs,
		Results:   results,
		Postcards: integrationPostcardSource{},
		Media:     media,
		Storage:   postcardaiapp.NoopStorage{},
		AI:        ai,
		Enabled:   true,
	})
	for i := 0; i < 2; i++ {
		if _, err := service.EnqueuePostcardAnalysis(context.Background(), postcardaiapp.EnqueueInput{PostcardID: 14, Reason: postcardaiapp.EnqueueReasonBackfill}); err != nil {
			t.Fatalf("enqueue %d: %v", i, err)
		}
		if err := appplatform.RunAnalysisWorker(context.Background(), service, appplatform.WorkerOptions{RunOnce: true}); err != nil {
			t.Fatalf("run worker %d: %v", i, err)
		}
	}
	if ai.imageCalls != 1 {
		t.Fatalf("image provider calls = %d, want 1", ai.imageCalls)
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
		if r.jobs[i].Status == domain.StatusPending && (r.jobs[i].NextRunAt == nil || !r.jobs[i].NextRunAt.After(now)) {
			r.jobs[i].Status = domain.StatusProcessing
			r.jobs[i].LockedAt = &now
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
	for i := range r.jobs {
		if r.jobs[i].ID == id {
			r.jobs[i].Status = domain.StatusPending
			r.jobs[i].Attempts++
			r.jobs[i].NextRunAt = &nextRunAt
			r.jobs[i].LastErrorCode = errorCode
		}
	}
	return nil
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
	media     map[uint]domain.MediaAnalysis
	postcards []domain.PostcardAnalysis
}

func newIntegrationResultRepo() *integrationResultRepo {
	return &integrationResultRepo{media: map[uint]domain.MediaAnalysis{}}
}

func (r *integrationResultRepo) FindReusableMediaAnalysis(ctx context.Context, key domain.VersionKey) (*domain.MediaAnalysis, error) {
	analysis, ok := r.media[key.ResourceID]
	if ok &&
		analysis.Status == domain.StatusSucceeded &&
		analysis.MediaVersion == key.ResourceVersion &&
		analysis.PromptVersion == key.PromptVersion &&
		analysis.SchemaVersion == key.SchemaVersion &&
		analysis.ModelVersion == key.ModelVersion {
		return &analysis, nil
	}
	return nil, nil
}

func (r *integrationResultRepo) StoreMediaAnalysis(ctx context.Context, analysis domain.MediaAnalysis) error {
	r.media[analysis.MediaID] = analysis
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

type integrationMediaSource struct {
	items []postcardaiapp.MediaSnapshot
}

func (s integrationMediaSource) ListMediaForPostcard(ctx context.Context, postcardID uint) ([]postcardaiapp.MediaSnapshot, error) {
	if s.items != nil {
		return s.items, nil
	}
	return []postcardaiapp.MediaSnapshot{{ID: 99, Version: "mv1", Type: "image", StorageKey: "media/99.jpg"}}, nil
}

type changingIntegrationPostcardSource struct {
	versions []string
	calls    int
}

func (s *changingIntegrationPostcardSource) GetPostcardForAnalysis(ctx context.Context, postcardID uint) (*postcardaiapp.PostcardSnapshot, error) {
	version := "pv1"
	if s.calls < len(s.versions) {
		version = s.versions[s.calls]
	}
	s.calls++
	return &postcardaiapp.PostcardSnapshot{ID: postcardID, Version: version, Content: json.RawMessage(`{"text":"lunch"}`), UpdatedAt: time.Now()}, nil
}

type unavailableIntegrationAI struct{}

func (unavailableIntegrationAI) AnalyzeImage(ctx context.Context, req postcardaiapp.ImageUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	return postcardaiapp.NoopAIClient{}.AnalyzeImage(ctx, req)
}

func (unavailableIntegrationAI) RepairImage(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return postcardaiapp.NoopAIClient{}.RepairImage(ctx, req)
}

func (unavailableIntegrationAI) AnalyzePostcard(ctx context.Context, req postcardaiapp.PostcardUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	return nil, domain.ProviderError{Code: domain.ErrorProviderUnavailable, Err: errors.New("temporary outage")}
}

func (unavailableIntegrationAI) RepairPostcard(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return postcardaiapp.NoopAIClient{}.RepairPostcard(ctx, req)
}

type countingIntegrationAI struct {
	imageCalls int
}

func (c *countingIntegrationAI) AnalyzeImage(ctx context.Context, req postcardaiapp.ImageUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	c.imageCalls++
	return postcardaiapp.NoopAIClient{}.AnalyzeImage(ctx, req)
}

func (c *countingIntegrationAI) RepairImage(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return postcardaiapp.NoopAIClient{}.RepairImage(ctx, req)
}

func (c *countingIntegrationAI) AnalyzePostcard(ctx context.Context, req postcardaiapp.PostcardUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	return postcardaiapp.NoopAIClient{}.AnalyzePostcard(ctx, req)
}

func (c *countingIntegrationAI) RepairPostcard(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return postcardaiapp.NoopAIClient{}.RepairPostcard(ctx, req)
}

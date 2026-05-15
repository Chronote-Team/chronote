package postcardai_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
)

type fakeJobRepository struct {
	nextID uint
	jobs   []domain.AnalysisJob
}

func newFakeJobRepository() *fakeJobRepository {
	return &fakeJobRepository{nextID: 1}
}

func (r *fakeJobRepository) Enqueue(ctx context.Context, job domain.AnalysisJob) (*domain.AnalysisJob, bool, error) {
	for i := range r.jobs {
		existing := &r.jobs[i]
		if existing.PostcardID == job.PostcardID && existing.PostcardVersion == job.PostcardVersion && existing.IsActive() {
			return existing, true, nil
		}
	}
	job.ID = r.nextID
	r.nextID++
	r.jobs = append(r.jobs, job)
	return &r.jobs[len(r.jobs)-1], false, nil
}

func (r *fakeJobRepository) ClaimNext(ctx context.Context, workerID string, now time.Time) (*domain.AnalysisJob, error) {
	for i := range r.jobs {
		if r.jobs[i].Status == domain.StatusPending {
			r.jobs[i].Status = domain.StatusProcessing
			r.jobs[i].LockedAt = &now
			return &r.jobs[i], nil
		}
	}
	return nil, nil
}

func (r *fakeJobRepository) UpdateStatus(ctx context.Context, id uint, status domain.AnalysisStatus, errorCode domain.ProviderErrorCode) error {
	for i := range r.jobs {
		if r.jobs[i].ID == id {
			r.jobs[i].Status = status
			r.jobs[i].LastErrorCode = errorCode
			return nil
		}
	}
	return fmt.Errorf("job %d not found", id)
}

func (r *fakeJobRepository) ScheduleRetry(ctx context.Context, id uint, nextRunAt time.Time, errorCode domain.ProviderErrorCode) error {
	for i := range r.jobs {
		if r.jobs[i].ID == id {
			r.jobs[i].Status = domain.StatusPending
			r.jobs[i].Attempts++
			r.jobs[i].NextRunAt = &nextRunAt
			r.jobs[i].LastErrorCode = errorCode
			return nil
		}
	}
	return fmt.Errorf("job %d not found", id)
}

func (r *fakeJobRepository) FindByID(ctx context.Context, id uint) (*domain.AnalysisJob, error) {
	for i := range r.jobs {
		if r.jobs[i].ID == id {
			return &r.jobs[i], nil
		}
	}
	return nil, nil
}

type fakeResultRepository struct {
	media     map[uint]domain.MediaAnalysis
	postcards []domain.PostcardAnalysis
}

func newFakeResultRepository() *fakeResultRepository {
	return &fakeResultRepository{media: map[uint]domain.MediaAnalysis{}}
}

func (r *fakeResultRepository) FindReusableMediaAnalysis(ctx context.Context, key domain.VersionKey) (*domain.MediaAnalysis, error) {
	analysis, ok := r.media[key.ResourceID]
	if !ok || analysis.Status != domain.StatusSucceeded {
		return nil, nil
	}
	if analysis.MediaVersion == key.ResourceVersion &&
		analysis.PromptVersion == key.PromptVersion &&
		analysis.SchemaVersion == key.SchemaVersion &&
		analysis.ModelVersion == key.ModelVersion {
		return &analysis, nil
	}
	return nil, nil
}

func (r *fakeResultRepository) StoreMediaAnalysis(ctx context.Context, analysis domain.MediaAnalysis) error {
	r.media[analysis.MediaID] = analysis
	return nil
}

func (r *fakeResultRepository) StorePostcardAnalysis(ctx context.Context, analysis domain.PostcardAnalysis) error {
	r.postcards = append(r.postcards, analysis)
	return nil
}

type fakePostcardSource struct {
	version string
	content json.RawMessage
}

func (s fakePostcardSource) GetPostcardForAnalysis(ctx context.Context, postcardID uint) (*postcardaiapp.PostcardSnapshot, error) {
	version := s.version
	if version == "" {
		version = "pv1"
	}
	content := s.content
	if len(content) == 0 {
		content = json.RawMessage(`{"blocks":[{"type":"text","value":"memory"}]}`)
	}
	return &postcardaiapp.PostcardSnapshot{ID: postcardID, Version: version, Content: content, UpdatedAt: time.Now()}, nil
}

type changingPostcardSource struct {
	versions []string
	calls    int
}

func (s *changingPostcardSource) GetPostcardForAnalysis(ctx context.Context, postcardID uint) (*postcardaiapp.PostcardSnapshot, error) {
	version := "current"
	if s.calls < len(s.versions) {
		version = s.versions[s.calls]
	}
	s.calls++
	return &postcardaiapp.PostcardSnapshot{ID: postcardID, Version: version, Content: json.RawMessage(`{"text":"hello"}`), UpdatedAt: time.Now()}, nil
}

type fakeMediaSource struct {
	items []postcardaiapp.MediaSnapshot
}

func (s fakeMediaSource) ListMediaForPostcard(ctx context.Context, postcardID uint) ([]postcardaiapp.MediaSnapshot, error) {
	return s.items, nil
}

type fakeStorage struct{}

func (fakeStorage) PresignGetObject(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	return "https://signed.example.com/" + objectKey, nil
}

type fakeAIClient struct{}

func (fakeAIClient) AnalyzeImage(ctx context.Context, req postcardaiapp.ImageUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	return &postcardaiapp.AIResult{JSON: json.RawMessage(`{"image_type":"food","caption":"ramen","confidence":0.9}`), Confidence: 0.9}, nil
}

func (fakeAIClient) RepairImage(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return &postcardaiapp.AIResult{JSON: json.RawMessage(`{"image_type":"food","caption":"ramen","confidence":0.8}`), Confidence: 0.8}, nil
}

func (fakeAIClient) AnalyzePostcard(ctx context.Context, req postcardaiapp.PostcardUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	return &postcardaiapp.AIResult{JSON: json.RawMessage(`{"summary":"lunch","suggested_title":"Lunch","confidence":0.9}`), Confidence: 0.9}, nil
}

func (fakeAIClient) RepairPostcard(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return &postcardaiapp.AIResult{JSON: json.RawMessage(`{"summary":"lunch","suggested_title":"Lunch","confidence":0.8}`), Confidence: 0.8}, nil
}

type countingAIClient struct {
	imageCalls int
}

func (c *countingAIClient) AnalyzeImage(ctx context.Context, req postcardaiapp.ImageUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	c.imageCalls++
	return fakeAIClient{}.AnalyzeImage(ctx, req)
}

func (c *countingAIClient) RepairImage(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return fakeAIClient{}.RepairImage(ctx, req)
}

func (c *countingAIClient) AnalyzePostcard(ctx context.Context, req postcardaiapp.PostcardUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	return fakeAIClient{}.AnalyzePostcard(ctx, req)
}

func (c *countingAIClient) RepairPostcard(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return fakeAIClient{}.RepairPostcard(ctx, req)
}

type repairAIClient struct {
	firstImage       json.RawMessage
	repairedImage    json.RawMessage
	postcard         json.RawMessage
	imageCalls       int
	repairImageCalls int
}

func (c *repairAIClient) AnalyzeImage(ctx context.Context, req postcardaiapp.ImageUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	c.imageCalls++
	return &postcardaiapp.AIResult{JSON: c.firstImage}, nil
}

func (c *repairAIClient) RepairImage(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	c.repairImageCalls++
	return &postcardaiapp.AIResult{JSON: c.repairedImage, Confidence: 0.8}, nil
}

func (c *repairAIClient) AnalyzePostcard(ctx context.Context, req postcardaiapp.PostcardUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	return &postcardaiapp.AIResult{JSON: c.postcard, Confidence: 0.8}, nil
}

func (c *repairAIClient) RepairPostcard(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return &postcardaiapp.AIResult{JSON: c.postcard, Confidence: 0.8}, nil
}

type partialAIClient struct{}

func (partialAIClient) AnalyzeImage(ctx context.Context, req postcardaiapp.ImageUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	if req.MediaID == 12 {
		return nil, domain.ProviderError{Code: domain.ErrorProviderRefused, Err: errors.New("refused")}
	}
	return fakeAIClient{}.AnalyzeImage(ctx, req)
}

func (partialAIClient) RepairImage(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return fakeAIClient{}.RepairImage(ctx, req)
}

func (partialAIClient) AnalyzePostcard(ctx context.Context, req postcardaiapp.PostcardUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	if !req.Partial {
		return nil, errors.New("expected partial final request")
	}
	return &postcardaiapp.AIResult{JSON: json.RawMessage(`{"summary":"trip","suggested_title":"Trip","confidence":0.6}`), Confidence: 0.6, Uncertainty: "one image unavailable"}, nil
}

func (partialAIClient) RepairPostcard(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return fakeAIClient{}.RepairPostcard(ctx, req)
}

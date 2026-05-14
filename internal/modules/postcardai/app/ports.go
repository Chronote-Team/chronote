package app

import (
	"context"
	"encoding/json"
	"time"

	"chronote-refactor/internal/modules/postcardai/domain"
)

const (
	DefaultImagePromptVersion    = "image_understanding_v1"
	DefaultImageSchemaVersion    = "image_understanding_v1"
	DefaultPostcardPromptVersion = "postcard_understanding_v1"
	DefaultPostcardSchemaVersion = "postcard_understanding_v1"
	DefaultModelVersion          = "test-model"
)

type JobRepository interface {
	Enqueue(ctx context.Context, job domain.AnalysisJob) (*domain.AnalysisJob, bool, error)
	ClaimNext(ctx context.Context, workerID string, now time.Time) (*domain.AnalysisJob, error)
	UpdateStatus(ctx context.Context, id uint, status domain.AnalysisStatus, errorCode domain.ProviderErrorCode) error
	ScheduleRetry(ctx context.Context, id uint, nextRunAt time.Time, errorCode domain.ProviderErrorCode) error
	FindByID(ctx context.Context, id uint) (*domain.AnalysisJob, error)
}

type ResultRepository interface {
	FindReusableMediaAnalysis(ctx context.Context, key domain.VersionKey) (*domain.MediaAnalysis, error)
	StoreMediaAnalysis(ctx context.Context, analysis domain.MediaAnalysis) error
	StorePostcardAnalysis(ctx context.Context, analysis domain.PostcardAnalysis) error
}

type PostcardSource interface {
	GetPostcardForAnalysis(ctx context.Context, postcardID uint) (*PostcardSnapshot, error)
}

type MediaSource interface {
	ListMediaForPostcard(ctx context.Context, postcardID uint) ([]MediaSnapshot, error)
}

type Storage interface {
	PresignGetObject(ctx context.Context, objectKey string, ttl time.Duration) (string, error)
}

type AIClient interface {
	AnalyzeImage(ctx context.Context, req ImageUnderstandingRequest) (*AIResult, error)
	RepairImage(ctx context.Context, req RepairRequest) (*AIResult, error)
	AnalyzePostcard(ctx context.Context, req PostcardUnderstandingRequest) (*AIResult, error)
	RepairPostcard(ctx context.Context, req RepairRequest) (*AIResult, error)
}

type Clock interface {
	Now() time.Time
}

type PostcardSnapshot struct {
	ID        uint
	Version   string
	Title     string
	Content   json.RawMessage
	UpdatedAt time.Time
}

type MediaSnapshot struct {
	ID         uint
	Version    string
	Type       string
	StorageKey string
	URL        string
	UpdatedAt  time.Time
}

type ImageUnderstandingRequest struct {
	MediaID       uint
	MediaVersion  string
	MediaType     string
	SignedURL     string
	PromptVersion string
	SchemaVersion string
	ModelVersion  string
}

type PostcardUnderstandingRequest struct {
	Postcard      PostcardSnapshot
	MediaAnalyses []domain.MediaAnalysis
	PromptVersion string
	SchemaVersion string
	ModelVersion  string
	Partial       bool
	Uncertainty   string
}

type RepairRequest struct {
	Original      json.RawMessage
	ValidationErr string
	SchemaVersion string
	ModelVersion  string
}

type AIResult struct {
	JSON        json.RawMessage
	Confidence  float64
	Uncertainty string
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

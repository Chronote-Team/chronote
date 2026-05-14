package app

import (
	"context"
	"errors"
	"time"

	"chronote-refactor/internal/modules/postcardai/domain"
)

type EnqueueReason string

const (
	EnqueueReasonCreate      EnqueueReason = "create"
	EnqueueReasonUpdate      EnqueueReason = "update"
	EnqueueReasonMediaChange EnqueueReason = "media_change"
	EnqueueReasonBackfill    EnqueueReason = "backfill"
	EnqueueReasonRetry       EnqueueReason = "retry"
)

type Dependencies struct {
	Jobs      JobRepository
	Results   ResultRepository
	Postcards PostcardSource
	Media     MediaSource
	Storage   Storage
	AI        AIClient
	Clock     Clock
	Enabled   bool
	Model     string
}

type Service struct {
	jobs      JobRepository
	results   ResultRepository
	postcards PostcardSource
	media     MediaSource
	storage   Storage
	ai        AIClient
	clock     Clock
	enabled   bool
	model     string
}

func NewService(deps Dependencies) *Service {
	clock := deps.Clock
	if clock == nil {
		clock = systemClock{}
	}
	model := deps.Model
	if model == "" {
		model = DefaultModelVersion
	}
	return &Service{
		jobs:      deps.Jobs,
		results:   deps.Results,
		postcards: deps.Postcards,
		media:     deps.Media,
		storage:   deps.Storage,
		ai:        deps.AI,
		clock:     clock,
		enabled:   deps.Enabled,
		model:     model,
	}
}

type EnqueueInput struct {
	PostcardID  uint
	Reason      EnqueueReason
	RequestedBy string
}

type EnqueueResult struct {
	Job      *domain.AnalysisJob
	Existing bool
	Noop     bool
}

func (s *Service) EnqueuePostcardAnalysis(ctx context.Context, input EnqueueInput) (*EnqueueResult, error) {
	if !s.enabled || s.jobs == nil || s.postcards == nil || input.PostcardID == 0 {
		return &EnqueueResult{Noop: true}, nil
	}
	snapshot, err := s.postcards.GetPostcardForAnalysis(ctx, input.PostcardID)
	if err != nil {
		return nil, err
	}
	if snapshot == nil || snapshot.Version == "" {
		return &EnqueueResult{Noop: true}, nil
	}
	job := domain.NewAnalysisJob(input.PostcardID, snapshot.Version, s.clock.Now())
	stored, existing, err := s.jobs.Enqueue(ctx, job)
	if err != nil {
		return nil, err
	}
	return &EnqueueResult{Job: stored, Existing: existing}, nil
}

func (s *Service) RetryAnalysisJob(ctx context.Context, jobID uint, reason string) (*domain.AnalysisJob, error) {
	if s.jobs == nil {
		return nil, errors.New("postcard ai jobs repository is not configured")
	}
	job, err := s.jobs.FindByID(ctx, jobID)
	if err != nil || job == nil {
		return job, err
	}
	if job.Status != domain.StatusFailed && job.Status != domain.StatusUnavailable {
		return nil, errors.New("analysis job is not retryable")
	}
	next := s.clock.Now()
	if err := s.jobs.ScheduleRetry(ctx, jobID, next, domain.ErrorNone); err != nil {
		return nil, err
	}
	return s.jobs.FindByID(ctx, jobID)
}

func (s *Service) BackfillPostcardAnalysis(ctx context.Context, postcardID uint) (*EnqueueResult, error) {
	return s.EnqueuePostcardAnalysis(ctx, EnqueueInput{PostcardID: postcardID, Reason: EnqueueReasonBackfill, RequestedBy: "system"})
}

func retryDelay(attempts int) time.Duration {
	if attempts <= 0 {
		return time.Second
	}
	if attempts > 6 {
		attempts = 6
	}
	return time.Duration(1<<attempts) * time.Second
}

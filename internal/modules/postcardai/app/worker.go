package app

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"chronote-refactor/internal/modules/postcardai/domain"
)

type WorkerOutcome struct {
	JobID  uint
	Status domain.AnalysisStatus
}

func (s *Service) RunNextAnalysisJob(ctx context.Context, workerID string) (*WorkerOutcome, error) {
	if !s.enabled || s.jobs == nil {
		return &WorkerOutcome{}, nil
	}
	now := s.clock.Now()
	job, err := s.jobs.ClaimNext(ctx, workerID, now)
	if err != nil || job == nil {
		return nil, err
	}

	postcard, err := s.postcards.GetPostcardForAnalysis(ctx, job.PostcardID)
	if err != nil {
		return nil, s.failJob(ctx, job, domain.ErrorProviderUnavailable)
	}
	if postcard == nil || !job.IsCurrentFor(postcard.Version) {
		return &WorkerOutcome{JobID: job.ID, Status: domain.StatusStale}, s.jobs.UpdateStatus(ctx, job.ID, domain.StatusStale, domain.ErrorStaleVersion)
	}

	medias, err := s.listMedia(ctx, job.PostcardID)
	if err != nil {
		return nil, s.failJob(ctx, job, domain.ErrorProviderUnavailable)
	}

	mediaAnalyses := make([]domain.MediaAnalysis, 0, len(medias))
	partial := false
	uncertainties := []string{}
	for _, media := range medias {
		if media.Type != "" && media.Type != "image" {
			continue
		}
		analysis, reused, err := s.mediaAnalysis(ctx, media)
		if err != nil {
			partial = true
			uncertainties = append(uncertainties, "media analysis unavailable")
			failed := failedMediaAnalysis(media, s.model, domain.ErrorCode(err), now)
			_ = s.results.StoreMediaAnalysis(ctx, failed)
			continue
		}
		if !reused {
			if err := s.results.StoreMediaAnalysis(ctx, *analysis); err != nil {
				return nil, err
			}
		}
		mediaAnalyses = append(mediaAnalyses, *analysis)
	}

	current, err := s.postcards.GetPostcardForAnalysis(ctx, job.PostcardID)
	if err != nil {
		return nil, s.failJob(ctx, job, domain.ErrorProviderUnavailable)
	}
	if current == nil || !job.IsCurrentFor(current.Version) {
		return &WorkerOutcome{JobID: job.ID, Status: domain.StatusStale}, s.jobs.UpdateStatus(ctx, job.ID, domain.StatusStale, domain.ErrorStaleVersion)
	}

	final, err := s.postcardAnalysis(ctx, *postcard, mediaAnalyses, partial, strings.Join(uncertainties, "; "))
	if err != nil {
		code := domain.ErrorCode(err)
		if domain.IsPermanent(code) {
			if updateErr := s.jobs.UpdateStatus(ctx, job.ID, domain.StatusUnavailable, code); updateErr != nil {
				return nil, updateErr
			}
			return &WorkerOutcome{JobID: job.ID, Status: domain.StatusUnavailable}, nil
		}
		return nil, s.failJob(ctx, job, code)
	}
	if err := s.results.StorePostcardAnalysis(ctx, *final); err != nil {
		return nil, err
	}
	if err := s.jobs.UpdateStatus(ctx, job.ID, domain.StatusSucceeded, domain.ErrorNone); err != nil {
		return nil, err
	}
	return &WorkerOutcome{JobID: job.ID, Status: domain.StatusSucceeded}, nil
}

func (s *Service) listMedia(ctx context.Context, postcardID uint) ([]MediaSnapshot, error) {
	if s.media == nil {
		return nil, nil
	}
	return s.media.ListMediaForPostcard(ctx, postcardID)
}

func (s *Service) mediaAnalysis(ctx context.Context, media MediaSnapshot) (*domain.MediaAnalysis, bool, error) {
	key := domain.VersionKey{
		ResourceID:      media.ID,
		ResourceVersion: media.Version,
		PromptVersion:   DefaultImagePromptVersion,
		SchemaVersion:   DefaultImageSchemaVersion,
		ModelVersion:    s.model,
	}
	if s.results != nil {
		existing, err := s.results.FindReusableMediaAnalysis(ctx, key)
		if err != nil {
			return nil, false, err
		}
		if existing != nil {
			return existing, true, nil
		}
	}
	if s.storage == nil || s.ai == nil {
		return nil, false, errors.New("postcard ai storage or provider is not configured")
	}
	signedURL, err := s.storage.PresignGetObject(ctx, media.StorageKey, 10*time.Minute)
	if err != nil {
		return nil, false, err
	}
	result, err := s.ai.AnalyzeImage(ctx, ImageUnderstandingRequest{
		MediaID:       media.ID,
		MediaVersion:  media.Version,
		MediaType:     media.Type,
		SignedURL:     signedURL,
		PromptVersion: DefaultImagePromptVersion,
		SchemaVersion: DefaultImageSchemaVersion,
		ModelVersion:  s.model,
	})
	if err != nil {
		return nil, false, err
	}
	result, err = s.validateOrRepairImage(ctx, result)
	if err != nil {
		return nil, false, err
	}
	now := s.clock.Now()
	return &domain.MediaAnalysis{
		MediaID:       media.ID,
		MediaVersion:  media.Version,
		PromptVersion: DefaultImagePromptVersion,
		SchemaVersion: DefaultImageSchemaVersion,
		ModelVersion:  s.model,
		Status:        domain.StatusSucceeded,
		Result:        cloneJSON(result.JSON),
		Confidence:    result.Confidence,
		Uncertainty:   result.Uncertainty,
		CreatedAt:     now,
		UpdatedAt:     now,
	}, false, nil
}

func (s *Service) validateOrRepairImage(ctx context.Context, result *AIResult) (*AIResult, error) {
	if result == nil {
		return nil, domain.ProviderError{Code: domain.ErrorMalformedOutput, Err: errors.New("empty image result")}
	}
	if err := validateImageUnderstanding(result.JSON); err == nil {
		return result, nil
	} else if s.ai != nil {
		repaired, repairErr := s.ai.RepairImage(ctx, RepairRequest{
			Original:      result.JSON,
			ValidationErr: err.Error(),
			SchemaVersion: DefaultImageSchemaVersion,
			ModelVersion:  s.model,
		})
		if repairErr != nil {
			return nil, repairErr
		}
		if validateErr := validateImageUnderstanding(repaired.JSON); validateErr != nil {
			return nil, domain.ProviderError{Code: domain.ErrorMalformedOutput, Err: validateErr}
		}
		return repaired, nil
	}
	return nil, domain.ProviderError{Code: domain.ErrorMalformedOutput}
}

func (s *Service) postcardAnalysis(ctx context.Context, postcard PostcardSnapshot, mediaAnalyses []domain.MediaAnalysis, partial bool, uncertainty string) (*domain.PostcardAnalysis, error) {
	if s.ai == nil {
		return nil, errors.New("postcard ai provider is not configured")
	}
	result, err := s.ai.AnalyzePostcard(ctx, PostcardUnderstandingRequest{
		Postcard:      postcard,
		MediaAnalyses: mediaAnalyses,
		PromptVersion: DefaultPostcardPromptVersion,
		SchemaVersion: DefaultPostcardSchemaVersion,
		ModelVersion:  s.model,
		Partial:       partial,
		Uncertainty:   uncertainty,
	})
	if err != nil {
		return nil, err
	}
	result, err = s.validateOrRepairPostcard(ctx, result)
	if err != nil {
		return nil, err
	}
	now := s.clock.Now()
	combinedUncertainty := strings.TrimSpace(strings.Join(nonEmpty(uncertainty, result.Uncertainty), "; "))
	return &domain.PostcardAnalysis{
		PostcardID:      postcard.ID,
		PostcardVersion: postcard.Version,
		PromptVersion:   DefaultPostcardPromptVersion,
		SchemaVersion:   DefaultPostcardSchemaVersion,
		ModelVersion:    s.model,
		Status:          domain.StatusSucceeded,
		Result:          cloneJSON(result.JSON),
		Confidence:      result.Confidence,
		Uncertainty:     combinedUncertainty,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (s *Service) validateOrRepairPostcard(ctx context.Context, result *AIResult) (*AIResult, error) {
	if result == nil {
		return nil, domain.ProviderError{Code: domain.ErrorMalformedOutput, Err: errors.New("empty postcard result")}
	}
	if err := validatePostcardUnderstanding(result.JSON); err == nil {
		return result, nil
	} else if s.ai != nil {
		repaired, repairErr := s.ai.RepairPostcard(ctx, RepairRequest{
			Original:      result.JSON,
			ValidationErr: err.Error(),
			SchemaVersion: DefaultPostcardSchemaVersion,
			ModelVersion:  s.model,
		})
		if repairErr != nil {
			return nil, repairErr
		}
		if validateErr := validatePostcardUnderstanding(repaired.JSON); validateErr != nil {
			return nil, domain.ProviderError{Code: domain.ErrorMalformedOutput, Err: validateErr}
		}
		return repaired, nil
	}
	return nil, domain.ProviderError{Code: domain.ErrorMalformedOutput}
}

func failedMediaAnalysis(media MediaSnapshot, model string, code domain.ProviderErrorCode, now time.Time) domain.MediaAnalysis {
	return domain.MediaAnalysis{
		MediaID:       media.ID,
		MediaVersion:  media.Version,
		PromptVersion: DefaultImagePromptVersion,
		SchemaVersion: DefaultImageSchemaVersion,
		ModelVersion:  model,
		Status:        domain.StatusUnavailable,
		ErrorCode:     code,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func (s *Service) failJob(ctx context.Context, job *domain.AnalysisJob, code domain.ProviderErrorCode) error {
	if job.Attempts < 1 && code == domain.ErrorProviderUnavailable {
		next := s.clock.Now().Add(retryDelay(job.Attempts))
		return s.jobs.ScheduleRetry(ctx, job.ID, next, code)
	}
	return s.jobs.UpdateStatus(ctx, job.ID, domain.StatusFailed, code)
}

func cloneJSON(raw json.RawMessage) json.RawMessage {
	if raw == nil {
		return nil
	}
	cloned := make([]byte, len(raw))
	copy(cloned, raw)
	return cloned
}

func nonEmpty(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, strings.TrimSpace(value))
		}
	}
	return out
}

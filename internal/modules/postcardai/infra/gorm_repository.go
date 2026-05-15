package infra

import (
	"context"
	"time"

	"chronote-refactor/internal/modules/postcardai/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Enqueue(ctx context.Context, job domain.AnalysisJob) (*domain.AnalysisJob, bool, error) {
	var existing AnalysisJobModel
	err := r.db.WithContext(ctx).
		Where("postcard_id = ? AND postcard_version = ? AND status IN ?", job.PostcardID, job.PostcardVersion, []string{string(domain.StatusPending), string(domain.StatusProcessing)}).
		First(&existing).Error
	if err == nil {
		mapped := jobFromModel(existing)
		return &mapped, true, nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, false, err
	}

	model := AnalysisJobModel{
		PostcardID:      job.PostcardID,
		PostcardVersion: job.PostcardVersion,
		Status:          string(job.Status),
		Attempts:        job.Attempts,
		NextRunAt:       job.NextRunAt,
		LockedAt:        job.LockedAt,
		LastErrorCode:   string(job.LastErrorCode),
		CreatedAt:       job.CreatedAt,
		UpdatedAt:       job.UpdatedAt,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, false, err
	}
	mapped := jobFromModel(model)
	return &mapped, false, nil
}

func (r *GormRepository) ClaimNext(ctx context.Context, workerID string, now time.Time) (*domain.AnalysisJob, error) {
	var model AnalysisJobModel
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND (next_run_at IS NULL OR next_run_at <= ?)", string(domain.StatusPending), now).
			Order("created_at ASC").
			First(&model).Error; err != nil {
			return err
		}
		return tx.Model(&model).Updates(map[string]any{
			"status":     string(domain.StatusProcessing),
			"locked_at":  now,
			"updated_at": now,
		}).Error
	})
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	model.Status = string(domain.StatusProcessing)
	model.LockedAt = &now
	mapped := jobFromModel(model)
	return &mapped, nil
}

func (r *GormRepository) UpdateStatus(ctx context.Context, id uint, status domain.AnalysisStatus, errorCode domain.ProviderErrorCode) error {
	return r.db.WithContext(ctx).Model(&AnalysisJobModel{}).Where("id = ?", id).Updates(map[string]any{
		"status":          string(status),
		"last_error_code": string(errorCode),
		"updated_at":      time.Now(),
	}).Error
}

func (r *GormRepository) ScheduleRetry(ctx context.Context, id uint, nextRunAt time.Time, errorCode domain.ProviderErrorCode) error {
	return r.db.WithContext(ctx).Model(&AnalysisJobModel{}).Where("id = ?", id).Updates(map[string]any{
		"status":          string(domain.StatusPending),
		"attempts":        gorm.Expr("attempts + 1"),
		"next_run_at":     nextRunAt,
		"last_error_code": string(errorCode),
		"updated_at":      time.Now(),
	}).Error
}

func (r *GormRepository) FindByID(ctx context.Context, id uint) (*domain.AnalysisJob, error) {
	var model AnalysisJobModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	mapped := jobFromModel(model)
	return &mapped, nil
}

func (r *GormRepository) FindReusableMediaAnalysis(ctx context.Context, key domain.VersionKey) (*domain.MediaAnalysis, error) {
	var model MediaAnalysisModel
	err := r.db.WithContext(ctx).
		Where("media_id = ? AND media_version = ? AND prompt_version = ? AND schema_version = ? AND model_version = ? AND status = ?",
			key.ResourceID, key.ResourceVersion, key.PromptVersion, key.SchemaVersion, key.ModelVersion, string(domain.StatusSucceeded)).
		First(&model).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	mapped := mediaFromModel(model)
	return &mapped, nil
}

func (r *GormRepository) StoreMediaAnalysis(ctx context.Context, analysis domain.MediaAnalysis) error {
	model := mediaToModel(analysis)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "media_id"}, {Name: "media_version"}, {Name: "prompt_version"}, {Name: "schema_version"}, {Name: "model_version"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"status", "result_json", "confidence", "uncertainty", "error_code", "updated_at"}),
	}).Create(&model).Error
}

func (r *GormRepository) StorePostcardAnalysis(ctx context.Context, analysis domain.PostcardAnalysis) error {
	model := postcardToModel(analysis)
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "postcard_id"}, {Name: "postcard_version"}, {Name: "prompt_version"}, {Name: "schema_version"}, {Name: "model_version"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"status", "result_json", "confidence", "uncertainty", "error_code", "updated_at"}),
	}).Create(&model).Error
}

func jobFromModel(model AnalysisJobModel) domain.AnalysisJob {
	return domain.AnalysisJob{
		ID:              model.ID,
		PostcardID:      model.PostcardID,
		PostcardVersion: model.PostcardVersion,
		Status:          domain.AnalysisStatus(model.Status),
		Attempts:        model.Attempts,
		NextRunAt:       model.NextRunAt,
		LockedAt:        model.LockedAt,
		LastErrorCode:   domain.ProviderErrorCode(model.LastErrorCode),
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}

func mediaToModel(analysis domain.MediaAnalysis) MediaAnalysisModel {
	return MediaAnalysisModel{
		ID:            analysis.ID,
		MediaID:       analysis.MediaID,
		MediaVersion:  analysis.MediaVersion,
		PromptVersion: analysis.PromptVersion,
		SchemaVersion: analysis.SchemaVersion,
		ModelVersion:  analysis.ModelVersion,
		Status:        string(analysis.Status),
		ResultJSON:    []byte(analysis.Result),
		Confidence:    analysis.Confidence,
		Uncertainty:   analysis.Uncertainty,
		ErrorCode:     string(analysis.ErrorCode),
		CreatedAt:     analysis.CreatedAt,
		UpdatedAt:     analysis.UpdatedAt,
	}
}

func mediaFromModel(model MediaAnalysisModel) domain.MediaAnalysis {
	return domain.MediaAnalysis{
		ID:            model.ID,
		MediaID:       model.MediaID,
		MediaVersion:  model.MediaVersion,
		PromptVersion: model.PromptVersion,
		SchemaVersion: model.SchemaVersion,
		ModelVersion:  model.ModelVersion,
		Status:        domain.AnalysisStatus(model.Status),
		Result:        append([]byte(nil), model.ResultJSON...),
		Confidence:    model.Confidence,
		Uncertainty:   model.Uncertainty,
		ErrorCode:     domain.ProviderErrorCode(model.ErrorCode),
		CreatedAt:     model.CreatedAt,
		UpdatedAt:     model.UpdatedAt,
	}
}

func postcardToModel(analysis domain.PostcardAnalysis) PostcardAnalysisModel {
	return PostcardAnalysisModel{
		ID:              analysis.ID,
		PostcardID:      analysis.PostcardID,
		PostcardVersion: analysis.PostcardVersion,
		PromptVersion:   analysis.PromptVersion,
		SchemaVersion:   analysis.SchemaVersion,
		ModelVersion:    analysis.ModelVersion,
		Status:          string(analysis.Status),
		ResultJSON:      []byte(analysis.Result),
		Confidence:      analysis.Confidence,
		Uncertainty:     analysis.Uncertainty,
		ErrorCode:       string(analysis.ErrorCode),
		CreatedAt:       analysis.CreatedAt,
		UpdatedAt:       analysis.UpdatedAt,
	}
}

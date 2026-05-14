package app

import (
	"context"
	"time"

	mediadomain "chronote-refactor/internal/modules/media/domain"
	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
)

type Repository interface {
	Create(media *mediadomain.Media) (*mediadomain.Media, error)
	ListByPostcardID(postcardID uint) ([]mediadomain.Media, error)
	CountByPostcardID(postcardID uint) (int, error)
	CountByPostcardIDAndGroup(postcardID uint, group string) (int, error)
	Reorder(postcardID uint, mediaIDs []uint) error
	Delete(postcardID, mediaID uint) error
	DeleteByPostcardID(postcardID uint) error
}

type AnalysisEnqueuer interface {
	EnqueuePostcardAnalysis(ctx context.Context, input postcardaiapp.EnqueueInput) (*postcardaiapp.EnqueueResult, error)
}

type SourceAdapter struct {
	repo Repository
}

func NewSourceAdapter(repo Repository) SourceAdapter {
	return SourceAdapter{repo: repo}
}

func (a SourceAdapter) ListMediaForPostcard(ctx context.Context, postcardID uint) ([]postcardaiapp.MediaSnapshot, error) {
	medias, err := a.repo.ListByPostcardID(postcardID)
	if err != nil {
		return nil, err
	}
	snapshots := make([]postcardaiapp.MediaSnapshot, 0, len(medias))
	for _, media := range medias {
		snapshots = append(snapshots, postcardaiapp.MediaSnapshot{
			ID:         media.ID,
			Version:    mediaVersion(media),
			Type:       media.Type,
			StorageKey: media.StorageKey,
			URL:        media.URL,
			UpdatedAt:  media.UpdatedAt,
		})
	}
	return snapshots, nil
}

func mediaVersion(media mediadomain.Media) string {
	updatedAt := media.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Unix(0, 0)
	}
	return updatedAt.UTC().Format(time.RFC3339Nano)
}

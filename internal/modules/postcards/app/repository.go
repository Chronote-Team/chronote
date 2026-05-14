package app

import (
	"context"
	"encoding/json"
	"time"

	mediadomain "chronote-refactor/internal/modules/media/domain"
	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	postcardsdomain "chronote-refactor/internal/modules/postcards/domain"
	usersdomain "chronote-refactor/internal/modules/users/domain"
)

type Repository interface {
	Create(postcard *postcardsdomain.Postcard) error
	FindByID(id uint) (*postcardsdomain.Postcard, error)
	FindRandomAccessible(userID uint) (*postcardsdomain.Postcard, error)
	List() ([]postcardsdomain.Postcard, error)
	Update(postcard *postcardsdomain.Postcard) error
	Delete(id uint) error
}

type AuthorRepository interface {
	FindByID(id uint) (*usersdomain.User, error)
}

type MediaRepository interface {
	ListByPostcardID(postcardID uint) ([]mediadomain.Media, error)
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

func (a SourceAdapter) GetPostcardForAnalysis(ctx context.Context, postcardID uint) (*postcardaiapp.PostcardSnapshot, error) {
	postcard, err := a.repo.FindByID(postcardID)
	if err != nil || postcard == nil {
		return nil, err
	}
	return &postcardaiapp.PostcardSnapshot{
		ID:        postcard.ID,
		Version:   postcardVersion(postcard),
		Title:     postcard.Title,
		Content:   cloneRaw(postcard.Content),
		UpdatedAt: postcard.UpdatedAt,
	}, nil
}

func postcardVersion(postcard *postcardsdomain.Postcard) string {
	if postcard == nil {
		return ""
	}
	updatedAt := postcard.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = time.Unix(0, 0)
	}
	return updatedAt.UTC().Format(time.RFC3339Nano)
}

func cloneRaw(raw json.RawMessage) json.RawMessage {
	if raw == nil {
		return nil
	}
	cloned := make([]byte, len(raw))
	copy(cloned, raw)
	return cloned
}

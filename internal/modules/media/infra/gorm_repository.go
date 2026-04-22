package infra

import (
	"errors"
	"time"

	mediaapp "chronote-refactor/internal/modules/media/app"
	mediadomain "chronote-refactor/internal/modules/media/domain"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

type mediaModel struct {
	ID                  uint   `gorm:"primaryKey"`
	PostcardID          uint   `gorm:"not null;index"`
	MediaType           string `gorm:"column:media_type;type:varchar(20);not null"`
	URL                 string `gorm:"type:varchar(500);not null"`
	ThumbnailURL        string `gorm:"type:varchar(500)"`
	StorageKey          string `gorm:"column:oss_key;type:varchar(500);not null"`
	ThumbnailStorageKey string `gorm:"column:thumbnail_oss_key;type:varchar(500)"`
	OriginalWidth       int
	OriginalHeight      int
	Duration            int
	FileSize            int64     `gorm:"not null"`
	Position            int       `gorm:"not null"`
	Group               string    `gorm:"column:media_group;type:varchar(50);not null"`
	CreatedAt           time.Time `gorm:"autoCreateTime"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime"`
}

func (mediaModel) TableName() string {
	return "postcard_media"
}

func NewGormRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

var _ mediaapp.Repository = (*Repository)(nil)

func (r *Repository) Create(media *mediadomain.Media) (*mediadomain.Media, error) {
	if r.db == nil {
		return nil, errors.New("database not initialized")
	}
	model := fromDomain(media)
	if err := r.db.Create(&model).Error; err != nil {
		return nil, err
	}
	result := toDomain(model)
	return &result, nil
}

func (r *Repository) ListByPostcardID(postcardID uint) ([]mediadomain.Media, error) {
	if r.db == nil {
		return nil, errors.New("database not initialized")
	}
	var models []mediaModel
	if err := r.db.Where("postcard_id = ?", postcardID).Order("position asc").Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]mediadomain.Media, 0, len(models))
	for _, model := range models {
		items = append(items, toDomain(model))
	}
	return items, nil
}

func (r *Repository) CountByPostcardID(postcardID uint) (int, error) {
	if r.db == nil {
		return 0, errors.New("database not initialized")
	}
	var count int64
	if err := r.db.Model(&mediaModel{}).Where("postcard_id = ?", postcardID).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *Repository) CountByPostcardIDAndGroup(postcardID uint, group string) (int, error) {
	if r.db == nil {
		return 0, errors.New("database not initialized")
	}
	var count int64
	if err := r.db.Model(&mediaModel{}).Where("postcard_id = ? AND media_group = ?", postcardID, group).Count(&count).Error; err != nil {
		return 0, err
	}
	return int(count), nil
}

func (r *Repository) Reorder(postcardID uint, mediaIDs []uint) error {
	if r.db == nil {
		return errors.New("database not initialized")
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		for index, id := range mediaIDs {
			if err := tx.Model(&mediaModel{}).Where("postcard_id = ? AND id = ?", postcardID, id).Update("position", index+1).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) Delete(postcardID, mediaID uint) error {
	if r.db == nil {
		return errors.New("database not initialized")
	}
	result := r.db.Where("postcard_id = ? AND id = ?", postcardID, mediaID).Delete(&mediaModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *Repository) DeleteByPostcardID(postcardID uint) error {
	if r.db == nil {
		return errors.New("database not initialized")
	}
	return r.db.Where("postcard_id = ?", postcardID).Delete(&mediaModel{}).Error
}

func fromDomain(media *mediadomain.Media) mediaModel {
	return mediaModel{
		ID:                  media.ID,
		PostcardID:          media.PostcardID,
		MediaType:           media.Type,
		URL:                 media.URL,
		ThumbnailURL:        media.ThumbnailURL,
		StorageKey:          media.StorageKey,
		ThumbnailStorageKey: media.ThumbnailStorageKey,
		OriginalWidth:       media.OriginalWidth,
		OriginalHeight:      media.OriginalHeight,
		Duration:            media.Duration,
		FileSize:            media.FileSize,
		Position:            media.Position,
		Group:               media.Group,
		CreatedAt:           media.CreatedAt,
		UpdatedAt:           media.UpdatedAt,
	}
}

func toDomain(model mediaModel) mediadomain.Media {
	return mediadomain.Media{
		ID:                  model.ID,
		PostcardID:          model.PostcardID,
		Type:                model.MediaType,
		URL:                 model.URL,
		ThumbnailURL:        model.ThumbnailURL,
		StorageKey:          model.StorageKey,
		ThumbnailStorageKey: model.ThumbnailStorageKey,
		OriginalWidth:       model.OriginalWidth,
		OriginalHeight:      model.OriginalHeight,
		Duration:            model.Duration,
		FileSize:            model.FileSize,
		Position:            model.Position,
		Group:               model.Group,
		CreatedAt:           model.CreatedAt,
		UpdatedAt:           model.UpdatedAt,
	}
}

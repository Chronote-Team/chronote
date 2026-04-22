package infra

import (
	"encoding/json"
	"errors"
	"time"

	postcardsapp "chronote-refactor/internal/modules/postcards/app"
	postcardsdomain "chronote-refactor/internal/modules/postcards/domain"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

type postcardModel struct {
	ID         uint      `gorm:"primaryKey"`
	Title      string    `gorm:"type:varchar(200);not null"`
	Content    []byte    `gorm:"type:jsonb;not null"`
	Visibility string    `gorm:"type:varchar(20);not null"`
	AuthorID   uint      `gorm:"not null;index"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}

func (postcardModel) TableName() string {
	return "postcards"
}

func NewGormRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

var _ postcardsapp.Repository = (*Repository)(nil)

func (r *Repository) Create(postcard *postcardsdomain.Postcard) error {
	if r.db == nil {
		return errors.New("database not initialized")
	}
	model := fromDomain(postcard)
	if err := r.db.Create(&model).Error; err != nil {
		return err
	}
	*postcard = toDomain(model)
	return nil
}

func (r *Repository) FindByID(id uint) (*postcardsdomain.Postcard, error) {
	if r.db == nil {
		return nil, errors.New("database not initialized")
	}
	var model postcardModel
	if err := r.db.First(&model, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	postcard := toDomain(model)
	return &postcard, nil
}

func (r *Repository) List() ([]postcardsdomain.Postcard, error) {
	if r.db == nil {
		return nil, errors.New("database not initialized")
	}
	var models []postcardModel
	if err := r.db.Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]postcardsdomain.Postcard, 0, len(models))
	for _, model := range models {
		items = append(items, toDomain(model))
	}
	return items, nil
}

func (r *Repository) Update(postcard *postcardsdomain.Postcard) error {
	if r.db == nil {
		return errors.New("database not initialized")
	}
	model := fromDomain(postcard)
	return r.db.Model(&postcardModel{}).Where("id = ?", postcard.ID).Updates(model).Error
}

func (r *Repository) Delete(id uint) error {
	if r.db == nil {
		return errors.New("database not initialized")
	}
	return r.db.Delete(&postcardModel{}, id).Error
}

func fromDomain(postcard *postcardsdomain.Postcard) postcardModel {
	return postcardModel{
		ID:         postcard.ID,
		Title:      postcard.Title,
		Content:    append([]byte(nil), postcard.Content...),
		Visibility: postcard.Visibility,
		AuthorID:   postcard.AuthorID,
		CreatedAt:  postcard.CreatedAt,
		UpdatedAt:  postcard.UpdatedAt,
	}
}

func toDomain(model postcardModel) postcardsdomain.Postcard {
	content := json.RawMessage(append([]byte(nil), model.Content...))
	return postcardsdomain.Postcard{
		ID:         model.ID,
		Title:      model.Title,
		Content:    content,
		Visibility: model.Visibility,
		AuthorID:   model.AuthorID,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}
}

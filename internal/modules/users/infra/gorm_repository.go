package infra

import (
	"errors"
	"time"

	usersapp "chronote-refactor/internal/modules/users/app"
	usersdomain "chronote-refactor/internal/modules/users/domain"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

type userModel struct {
	ID           uint      `gorm:"primaryKey"`
	Username     string    `gorm:"type:varchar(50);uniqueIndex;not null"`
	DisplayName  string    `gorm:"type:varchar(100)"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null"`
	PasswordHash string    `gorm:"column:password;type:varchar(255);not null"`
	Avatar       string    `gorm:"type:varchar(500)"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (userModel) TableName() string {
	return "users"
}

func NewGormRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

var _ usersapp.Repository = (*Repository)(nil)

func (r *Repository) Create(user *usersdomain.User) error {
	if r.db == nil {
		return errors.New("database not initialized")
	}
	model := fromDomain(user)
	if err := r.db.Create(&model).Error; err != nil {
		return err
	}
	*user = toDomain(model)
	return nil
}

func (r *Repository) FindByID(id uint) (*usersdomain.User, error) {
	return r.find("id = ?", id)
}

func (r *Repository) FindByEmail(email string) (*usersdomain.User, error) {
	return r.find("email = ?", email)
}

func (r *Repository) FindByUsername(username string) (*usersdomain.User, error) {
	return r.find("username = ?", username)
}

func (r *Repository) Update(user *usersdomain.User) error {
	if r.db == nil {
		return errors.New("database not initialized")
	}
	model := fromDomain(user)
	return r.db.Model(&userModel{}).Where("id = ?", user.ID).Updates(model).Error
}

func (r *Repository) find(query string, value interface{}) (*usersdomain.User, error) {
	if r.db == nil {
		return nil, errors.New("database not initialized")
	}
	var model userModel
	if err := r.db.Where(query, value).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	user := toDomain(model)
	return &user, nil
}

func fromDomain(user *usersdomain.User) userModel {
	return userModel{
		ID:           user.ID,
		Username:     user.Username,
		DisplayName:  user.DisplayName,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Avatar:       user.Avatar,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func toDomain(model userModel) usersdomain.User {
	return usersdomain.User{
		ID:           model.ID,
		Username:     model.Username,
		DisplayName:  model.DisplayName,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		Avatar:       model.Avatar,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

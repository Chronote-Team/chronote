package app

import (
	mediadomain "chronote-refactor/internal/modules/media/domain"
	postcardsdomain "chronote-refactor/internal/modules/postcards/domain"
	usersdomain "chronote-refactor/internal/modules/users/domain"
)

type Repository interface {
	Create(postcard *postcardsdomain.Postcard) error
	FindByID(id uint) (*postcardsdomain.Postcard, error)
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

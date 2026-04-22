package app

import mediadomain "chronote-refactor/internal/modules/media/domain"

type Repository interface {
	Create(media *mediadomain.Media) (*mediadomain.Media, error)
	ListByPostcardID(postcardID uint) ([]mediadomain.Media, error)
	CountByPostcardID(postcardID uint) (int, error)
	CountByPostcardIDAndGroup(postcardID uint, group string) (int, error)
	Reorder(postcardID uint, mediaIDs []uint) error
	Delete(postcardID, mediaID uint) error
	DeleteByPostcardID(postcardID uint) error
}

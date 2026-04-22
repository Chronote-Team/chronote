package app

import usersdomain "chronote-refactor/internal/modules/users/domain"

type Repository interface {
	Create(user *usersdomain.User) error
	FindByID(id uint) (*usersdomain.User, error)
	FindByEmail(email string) (*usersdomain.User, error)
	FindByUsername(username string) (*usersdomain.User, error)
	Update(user *usersdomain.User) error
}

type PasswordManager interface {
	Hash(string) (string, error)
	Verify(string, string) (bool, error)
}

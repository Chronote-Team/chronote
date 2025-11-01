package services

import (
	"chronote/global"
	"chronote/models"
	"chronote/utils"
)

type UserService struct{}

func (s *UserService) Register(user *models.User) error {
	hashedPassword, err := utils.EncryptPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword
	if err := global.Db.Create(user).Error; err != nil {
		return err
	}
	return nil
}

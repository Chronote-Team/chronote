package services

import (
	"chronote/config"
	"chronote/global"
	"chronote/models"
	"chronote/utils"
	"errors"
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

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func (s *UserService) Login(email, password string) (*LoginResponse, error) {
	var user models.User

	// Verify User Email and Password
	if err := global.Db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在或密码错误")
	}
	passwordMatch, err := utils.VerifyPassword(password, user.Password)
	if err != nil || !passwordMatch {
		return nil, errors.New("用户不存在或密码错误")
	}

	// Generate Tokens
	accessToken, refreshToken, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, errors.New("Token 生成失败")
	}
	expiresInSeconds := config.AppConfig.JWT.AccessTokenExpire

	// Generate Login Response
	loginResponse := &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresInSeconds,
	}
	return loginResponse, nil
}

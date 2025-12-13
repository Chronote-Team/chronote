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

// UserInfoResponse represents the response structure for user info
type UserInfoResponse struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// GetUserInfo retrieves user information by user ID
func (s *UserService) GetUserInfo(userID uint) (*UserInfoResponse, error) {
	var user models.User
	if err := global.Db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	return &UserInfoResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
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
		return nil, errors.New("token 生成失败")
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

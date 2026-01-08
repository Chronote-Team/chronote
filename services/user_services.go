package services

import (
	"chronote/config"
	"chronote/global"
	"chronote/models"
	"chronote/utils"
	"context"
	"errors"
	"mime/multipart"
)

type UserService struct{}

func (s *UserService) Register(user *models.User) error {
	// 如果 display_name 为空，默认使用 username
	if user.DisplayName == "" {
		user.DisplayName = user.Username
	}
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
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	DisplayName string `json:"display_name"`
	Email      string `json:"email"`
	Avatar     string `json:"avatar,omitempty"`
	CreatedAt  string `json:"created_at"`
}

// GetUserInfo retrieves user information by user ID
func (s *UserService) GetUserInfo(userID uint) (*UserInfoResponse, error) {
	var user models.User
	if err := global.Db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	return &UserInfoResponse{
		ID:         user.ID,
		Username:   user.Username,
		DisplayName: user.DisplayName,
		Email:      user.Email,
		Avatar:     user.Avatar,
		CreatedAt:  user.CreatedAt.Format("2006-01-02 15:04:05"),
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

// RefreshTokenResponse represents the response structure for token refresh
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// isTokenBlacklisted checks if a token is in the Redis blacklist
func isTokenBlacklisted(token string) (bool, error) {
	tokenBlacklistService := TokenBlacklistService{}
	return tokenBlacklistService.IsBlacklisted(context.Background(), token)
}

// RefreshToken validates refresh token and generates new token pair
func (s *UserService) RefreshToken(refreshTokenString string) (*RefreshTokenResponse, error) {
	// Check if refresh token is blacklisted
	blacklisted, err := isTokenBlacklisted(refreshTokenString)
	if err != nil {
		return nil, errors.New("token 验证失败")
	}
	if blacklisted {
		return nil, errors.New("refresh token 已被撤销")
	}

	// Parse and validate refresh token
	claims, err := utils.Parsetoken(refreshTokenString)
	if err != nil {
		return nil, errors.New("refresh token 无效或已过期")
	}

	// Verify token type
	if claims.TokenType != "refresh" {
		return nil, errors.New("需要使用 refresh token")
	}

	// Verify user exists
	var user models.User
	if err := global.Db.First(&user, claims.UserID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	// Generate new token pair
	accessToken, refreshToken, err := utils.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, errors.New("token 生成失败")
	}

	return &RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    config.AppConfig.JWT.AccessTokenExpire,
	}, nil
}

// UpdateAvatar uploads a new avatar for the user and updates the database
func (s *UserService) UpdateAvatar(userID uint, file *multipart.FileHeader) (string, error) {
	// Get current user to check for existing avatar
	var user models.User
	if err := global.Db.First(&user, userID).Error; err != nil {
		return "", errors.New("用户不存在")
	}

	// Upload new avatar to S3
	avatarURL, err := utils.UploadAvatar(file, userID)
	if err != nil {
		return "", err
	}

	// Delete old avatar if exists
	if user.Avatar != "" {
		// Don't fail the request if old avatar deletion fails
		_ = utils.DeleteAvatar(user.Avatar)
	}

	// Update user avatar in database
	if err := global.Db.Model(&user).Update("avatar", avatarURL).Error; err != nil {
		// If database update fails, try to delete the newly uploaded avatar
		_ = utils.DeleteAvatar(avatarURL)
		return "", errors.New("更新用户头像失败")
	}

	return avatarURL, nil
}

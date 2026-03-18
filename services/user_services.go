package services

import (
	"chronote/config"
	"chronote/dto"
	"chronote/global"
	"chronote/models"
	"chronote/utils"
	"context"
	"errors"
	"mime/multipart"
	"regexp"
	"strings"
)

type UserService struct{}

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

const (
	minPasswordLength    = 6
	maxPasswordLength    = 72
	maxUsernameLength    = 50
	maxDisplayNameLength = 100
	maxEmailLength       = 255
)

func (s *UserService) Register(req *dto.RegisterRequest) (*models.User, error) {
	username := strings.TrimSpace(req.Username)
	displayName := strings.TrimSpace(req.DisplayName)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := req.Password

	if username == "" {
		return nil, errors.New("username 不能为空")
	}
	if len(username) > maxUsernameLength {
		return nil, errors.New("username 长度不能超过 50 个字符")
	}
	if !usernamePattern.MatchString(username) {
		return nil, errors.New("username 只能包含字母、数字和下划线")
	}
	if displayName != "" && len(displayName) > maxDisplayNameLength {
		return nil, errors.New("display_name 长度不能超过 100 个字符")
	}
	if email == "" {
		return nil, errors.New("email 不能为空")
	}
	if len(email) > maxEmailLength {
		return nil, errors.New("email 长度不能超过 255 个字符")
	}
	if strings.TrimSpace(password) == "" {
		return nil, errors.New("password 不能为空")
	}
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return nil, errors.New("password 长度必须在 6 到 72 个字符之间")
	}
	if displayName == "" {
		displayName = username
	}

	var count int64
	if err := global.Db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return nil, errors.New("用户注册失败")
	}
	if count > 0 {
		return nil, errors.New("username 已存在")
	}
	if err := global.Db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return nil, errors.New("用户注册失败")
	}
	if count > 0 {
		return nil, errors.New("email 已被使用")
	}

	hashedPassword, err := utils.EncryptPassword(password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}
	user := &models.User{
		Username:    username,
		DisplayName: displayName,
		Email:       email,
		Password:    hashedPassword,
	}
	if err := global.Db.Create(user).Error; err != nil {
		return nil, errors.New("用户注册失败")
	}
	return user, nil
}

// GetUserInfo retrieves user information by user ID
func (s *UserService) GetUserInfo(userID uint) (*dto.UserInfoResponse, error) {
	var user models.User
	if err := global.Db.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}

	response := dto.NewUserInfoResponse(&user)
	return &response, nil
}

func (s *UserService) Login(email, password string) (*dto.AuthTokenResponse, error) {
	var user models.User
	email = strings.ToLower(strings.TrimSpace(email))

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
	loginResponse := &dto.AuthTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresInSeconds,
	}
	return loginResponse, nil
}

// isTokenBlacklisted checks if a token is in the Redis blacklist
func isTokenBlacklisted(token string) (bool, error) {
	tokenBlacklistService := TokenBlacklistService{}
	return tokenBlacklistService.IsBlacklisted(context.Background(), token)
}

// RefreshToken validates refresh token and generates new token pair
func (s *UserService) RefreshToken(refreshTokenString string) (*dto.AuthTokenResponse, error) {
	refreshTokenString = strings.TrimSpace(refreshTokenString)
	if refreshTokenString == "" {
		return nil, errors.New("refresh_token 不能为空")
	}

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

	return &dto.AuthTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    config.AppConfig.JWT.AccessTokenExpire,
	}, nil
}

func (s *UserService) ValidateRefreshTokenForUser(userID uint, refreshTokenString string) error {
	refreshTokenString = strings.TrimSpace(refreshTokenString)
	if refreshTokenString == "" {
		return errors.New("refresh_token 不能为空")
	}

	claims, err := utils.Parsetoken(refreshTokenString)
	if err != nil {
		return errors.New("refresh token 无效或已过期")
	}
	if claims.TokenType != "refresh" {
		return errors.New("需要使用 refresh token")
	}
	if claims.UserID != userID {
		return errors.New("refresh token 不属于当前用户")
	}
	return nil
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

// UpdateDisplayName updates the user's display name
func (s *UserService) UpdateDisplayName(userID uint, displayName string) error {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return errors.New("display_name 不能为空")
	}
	if len(displayName) > maxDisplayNameLength {
		return errors.New("display_name 长度不能超过 100 个字符")
	}
	var user models.User
	if err := global.Db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}
	if err := global.Db.Model(&user).Update("display_name", displayName).Error; err != nil {
		return errors.New("更新显示名称失败")
	}
	return nil
}

// UpdatePassword updates the user's password after verifying the old password
func (s *UserService) UpdatePassword(userID uint, oldPassword, newPassword string) error {
	var user models.User
	if err := global.Db.First(&user, userID).Error; err != nil {
		return errors.New("用户不存在")
	}
	if strings.TrimSpace(newPassword) == "" {
		return errors.New("new_password 不能为空")
	}
	if len(newPassword) < minPasswordLength || len(newPassword) > maxPasswordLength {
		return errors.New("new_password 长度必须在 6 到 72 个字符之间")
	}
	if oldPassword == newPassword {
		return errors.New("新密码不能与旧密码相同")
	}

	// 验证旧密码
	passwordMatch, err := utils.VerifyPassword(oldPassword, user.Password)
	if err != nil || !passwordMatch {
		return errors.New("旧密码错误")
	}

	// 加密新密码
	hashedPassword, err := utils.EncryptPassword(newPassword)
	if err != nil {
		return errors.New("密码加密失败")
	}

	// 更新密码
	if err := global.Db.Model(&user).Update("password", hashedPassword).Error; err != nil {
		return errors.New("更新密码失败")
	}

	return nil
}

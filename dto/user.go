package dto

import "chronote/models"

type RegisterRequest struct {
	Username    string `json:"username" binding:"required,max=50"`
	DisplayName string `json:"display_name" binding:"omitempty,max=100"`
	Email       string `json:"email" binding:"required,email,max=255"`
	Password    string `json:"password" binding:"required,min=6,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateDisplayNameRequest struct {
	DisplayName string `json:"display_name" binding:"required,max=100"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=72"`
}

type RegisterUserResponse struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type UserInfoResponse struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Avatar      string `json:"avatar,omitempty"`
	CreatedAt   string `json:"created_at"`
}

type AuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type AvatarUploadResponse struct {
	AvatarURL string `json:"avatar_url"`
}

func NewRegisterUserResponse(user *models.User) RegisterUserResponse {
	return RegisterUserResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
	}
}

func NewUserInfoResponse(user *models.User) UserInfoResponse {
	return UserInfoResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Avatar:      user.Avatar,
		CreatedAt:   user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

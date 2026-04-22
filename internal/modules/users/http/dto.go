package http

import (
	"time"

	usersdomain "chronote-refactor/internal/modules/users/domain"
)

type RegisterRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Password    string `json:"password"`
}

type UpdateDisplayNameRequest struct {
	DisplayName string `json:"display_name"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
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

type AvatarUploadResponse struct {
	AvatarURL string `json:"avatar_url"`
}

func newRegisterUserResponse(user *usersdomain.User) RegisterUserResponse {
	return RegisterUserResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
	}
}

func newUserInfoResponse(user *usersdomain.User) UserInfoResponse {
	return UserInfoResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		Avatar:      user.Avatar,
		CreatedAt:   user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func avatarURL(userID uint, filename string) string {
	if filename == "" {
		return ""
	}
	return "/media/avatars/" + time.Now().Format("20060102150405") + "/" + filename
}

package http

import (
	"net/http"

	usersapp "chronote-refactor/internal/modules/users/app"
	"chronote-refactor/internal/shared/errs"
	"chronote-refactor/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *usersapp.Service
}

func NewHandler(service *usersapp.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Write(ctx, http.StatusBadRequest, "请求参数无效", nil)
		return
	}

	user, err := h.service.Register(usersapp.RegisterInput{
		Username:    req.Username,
		DisplayName: req.DisplayName,
		Email:       req.Email,
		Password:    req.Password,
	})
	if err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}

	response.Write(ctx, http.StatusCreated, "用户注册成功", gin.H{
		"user": newRegisterUserResponse(user),
	})
}

func (h *Handler) UserInfo(ctx *gin.Context) {
	userID, ok := ctx.Get("userID")
	if !ok {
		response.Write(ctx, http.StatusUnauthorized, "未授权访问", nil)
		return
	}

	user, err := h.service.GetUserInfo(userID.(uint))
	if err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}

	response.Write(ctx, http.StatusOK, "获取用户信息成功", newUserInfoResponse(user))
}

func (h *Handler) UploadAvatar(ctx *gin.Context) {
	userID, ok := ctx.Get("userID")
	if !ok {
		response.Write(ctx, http.StatusUnauthorized, "未授权访问", nil)
		return
	}
	file, err := ctx.FormFile("avatar")
	if err != nil {
		response.Write(ctx, http.StatusBadRequest, "请上传头像文件", nil)
		return
	}

	url := avatarURL(userID.(uint), file.Filename)
	if err := h.service.UpdateAvatar(userID.(uint), url); err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}

	response.Write(ctx, http.StatusOK, "头像上传成功", AvatarUploadResponse{AvatarURL: url})
}

func (h *Handler) UpdateDisplayName(ctx *gin.Context) {
	userID, ok := ctx.Get("userID")
	if !ok {
		response.Write(ctx, http.StatusUnauthorized, "未授权访问", nil)
		return
	}
	var req UpdateDisplayNameRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Write(ctx, http.StatusBadRequest, "请求参数无效", nil)
		return
	}
	if err := h.service.UpdateDisplayName(userID.(uint), req.DisplayName); err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	response.Write(ctx, http.StatusOK, "显示名称更新成功", nil)
}

func (h *Handler) UpdatePassword(ctx *gin.Context) {
	userID, ok := ctx.Get("userID")
	if !ok {
		response.Write(ctx, http.StatusUnauthorized, "未授权访问", nil)
		return
	}
	var req UpdatePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Write(ctx, http.StatusBadRequest, "请求参数无效", nil)
		return
	}
	if err := h.service.UpdatePassword(userID.(uint), req.OldPassword, req.NewPassword); err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	response.Write(ctx, http.StatusOK, "密码更新成功", nil)
}

package http

import (
	"net/http"
	"strings"

	authapp "chronote-refactor/internal/modules/auth/app"
	"chronote-refactor/internal/shared/errs"
	"chronote-refactor/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *authapp.Service
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func NewHandler(service *authapp.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Write(ctx, http.StatusBadRequest, "请求参数无效", nil)
		return
	}
	tokens, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	response.Write(ctx, http.StatusOK, "用户登录成功", AuthTokenResponse(*tokens))
}

func (h *Handler) RefreshToken(ctx *gin.Context) {
	var req RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Write(ctx, http.StatusBadRequest, "请求参数无效，需要提供 refresh_token", nil)
		return
	}
	tokens, err := h.service.RefreshToken(req.RefreshToken)
	if err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	response.Write(ctx, http.StatusOK, "Token 刷新成功", AuthTokenResponse(*tokens))
}

func (h *Handler) Logout(ctx *gin.Context) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		response.Write(ctx, http.StatusUnauthorized, "缺少 Authorization 头", nil)
		return
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		response.Write(ctx, http.StatusUnauthorized, "Token 格式无效", nil)
		return
	}
	var req LogoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Write(ctx, http.StatusBadRequest, "请求参数无效，需要提供 refresh_token", nil)
		return
	}
	userID, ok := ctx.Get("userID")
	if !ok {
		response.Write(ctx, http.StatusUnauthorized, "未授权访问", nil)
		return
	}
	if err := h.service.Logout(userID.(uint), parts[1], req.RefreshToken); err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	response.Write(ctx, http.StatusOK, "登出成功", nil)
}

package controllers

import (
	"chronote/models"
	"chronote/services"
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var userService = services.UserService{}
var tokenBlacklistService = services.TokenBlacklistService{}

// User Register Controller
func Register(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "请求参数无效"})
		return
	}
	if err := userService.Register(&user); err != nil {
		log.Printf("Failed to register user: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": "用户注册失败"})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"code":    http.StatusCreated,
		"message": "用户注册成功",
		"data": gin.H{
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
		},
	},
	)
}

// UserInfo retrieves the authenticated user's information
func UserInfo(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权访问",
		})
		return
	}

	userInfo, err := userService.GetUserInfo(userID.(uint))
	if err != nil {
		log.Printf("Failed to get user info: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"code":    http.StatusNotFound,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "获取用户信息成功",
		"data":    userInfo,
	})
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// User Login Controller
func Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "请求参数无效"})
		return
	}
	loginResponse, err := userService.Login(req.Email, req.Password)
	if err != nil {
		log.Printf("Failed to Login user: %v", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "用户登录成功",
		"data":    loginResponse,
	})
}

// RefreshToken refreshes access token using refresh token
func RefreshToken(ctx *gin.Context) {
	// Extract token from Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "缺少 Authorization 头",
		})
		return
	}

	// Validate Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "Token 格式无效",
		})
		return
	}

	refreshTokenResponse, err := userService.RefreshToken(parts[1])
	if err != nil {
		log.Printf("Failed to refresh token: %v", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "Token 刷新成功",
		"data":    refreshTokenResponse,
	})
}

// LogoutRequest represents the logout request body
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Logout handles user logout by blacklisting tokens
func Logout(ctx *gin.Context) {
	// Get access token from Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "缺少 Authorization 头",
		})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "Token 格式无效",
		})
		return
	}
	accessToken := parts[1]

	// Get refresh token from request body
	var req LogoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请求参数无效，需要提供 refresh_token",
		})
		return
	}

	// Add both tokens to blacklist
	if err := tokenBlacklistService.BlacklistTokenPair(context.Background(), accessToken, req.RefreshToken); err != nil {
		log.Printf("Failed to blacklist tokens: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": "登出失败",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "登出成功",
	})
}

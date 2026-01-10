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

// Register godoc
// @Summary      Register new user
// @Description  Create a new user account with username, email, and password
// @Tags         User,Public
// @Accept       json
// @Produce      json
// @Param        request body models.User true "User registration details"
// @Success      201  {object}  map[string]interface{}  "code,message,data.user"
// @Failure      400  {object}  map[string]interface{}  "code,message"
// @Failure      500  {object}  map[string]interface{}  "code,message"
// @Router       /user/register [post]
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
				"id":          user.ID,
				"username":    user.Username,
				"display_name": user.DisplayName,
				"email":       user.Email,
			},
		},
	},
	)
}

// UserInfo godoc
// @Summary      Get user information
// @Description  Retrieve the authenticated user's information
// @Tags         User,Protected
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "code,message,data"
// @Failure      401  {object}  map[string]interface{}  "code,message"
// @Failure      404  {object}  map[string]interface{}  "code,message"
// @Router       /user/info [get]
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
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// Login godoc
// @Summary      User login
// @Description  Authenticate user with email and password, returns JWT tokens
// @Tags         User,Public
// @Accept       json
// @Produce      json
// @Param        request body controllers.LoginRequest true "Login credentials"
// @Success      200  {object}  map[string]interface{}  "code,message.data{access_token,refresh_token,user}"
// @Failure      400  {object}  map[string]interface{}  "code,message"
// @Failure      401  {object}  map[string]interface{}  "code,message"
// @Router       /user/login [post]
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

// RefreshToken godoc
// @Summary      Refresh access token
// @Description  Get a new access token using a valid refresh token
// @Tags         User,Public
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "code,message.data{access_token,refresh_token}"
// @Failure      401  {object}  map[string]interface{}  "code,message"
// @Router       /user/refresh [post]
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
	RefreshToken string `json:"refresh_token" binding:"required" example:"your-refresh-token"`
}

// Logout godoc
// @Summary      User logout
// @Description  Logout user and blacklist both access and refresh tokens
// @Tags         User,Protected
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body controllers.LogoutRequest true "Refresh token to blacklist"
// @Success      200  {object}  map[string]interface{}  "code,message"
// @Failure      400  {object}  map[string]interface{}  "code,message"
// @Failure      401  {object}  map[string]interface{}  "code,message"
// @Failure      500  {object}  map[string]interface{}  "code,message"
// @Router       /user/logout [post]
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

// UploadAvatar godoc
// @Summary      Upload user avatar
// @Description  Upload or update the authenticated user's avatar image
// @Tags         User,Protected
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        avatar  formData  file  true  "Avatar image file (jpg, jpeg, png, gif, webp, max 2MB)"
// @Success      200  {object}  map[string]interface{}  "code,message.data.avatar_url"
// @Failure      400  {object}  map[string]interface{}  "code,message"
// @Failure      401  {object}  map[string]interface{}  "code,message"
// @Failure      500  {object}  map[string]interface{}  "code,message"
// @Router       /user/avatar [post]
func UploadAvatar(ctx *gin.Context) {
	// Get user ID from JWT middleware context
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权访问",
		})
		return
	}

	// Get uploaded file from form data
	file, err := ctx.FormFile("avatar")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请上传头像文件",
		})
		return
	}

	// Upload avatar via service layer
	avatarURL, err := userService.UpdateAvatar(userID.(uint), file)
	if err != nil {
		log.Printf("Failed to upload avatar: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "头像上传成功",
		"data": gin.H{
			"avatar_url": avatarURL,
		},
	})
}

// UpdateDisplayNameRequest represents the request for updating display name
type UpdateDisplayNameRequest struct {
	DisplayName string `json:"display_name" binding:"required" example:"John Doe"`
}

// UpdateDisplayName godoc
// @Summary      Update display name
// @Description  Update the authenticated user's display name
// @Tags         User,Protected
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body controllers.UpdateDisplayNameRequest true "New display name"
// @Success      200  {object}  map[string]interface{}  "code,message"
// @Failure      400  {object}  map[string]interface{}  "code,message"
// @Failure      401  {object}  map[string]interface{}  "code,message"
// @Failure      500  {object}  map[string]interface{}  "code,message"
// @Router       /user/update/displayname [put]
func UpdateDisplayName(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权访问",
		})
		return
	}

	var req UpdateDisplayNameRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请求参数无效",
		})
		return
	}

	if err := userService.UpdateDisplayName(userID.(uint), req.DisplayName); err != nil {
		log.Printf("Failed to update display name: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "显示名称更新成功",
	})
}

// UpdatePasswordRequest represents the request for updating password
type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required" example:"oldpassword123"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
}

// UpdatePassword godoc
// @Summary      Update password
// @Description  Update the authenticated user's password (requires old password verification)
// @Tags         User,Protected
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body controllers.UpdatePasswordRequest true "Old and new password"
// @Success      200  {object}  map[string]interface{}  "code,message"
// @Failure      400  {object}  map[string]interface{}  "code,message"
// @Failure      401  {object}  map[string]interface{}  "code,message"
// @Router       /user/update/password [put]
func UpdatePassword(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权访问",
		})
		return
	}

	var req UpdatePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请求参数无效",
		})
		return
	}

	if err := userService.UpdatePassword(userID.(uint), req.OldPassword, req.NewPassword); err != nil {
		log.Printf("Failed to update password: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "密码更新成功",
	})
}

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

func Register(ctx *gin.Context) {
	var req models.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "请求参数无效"})
		return
	}
	user, err := userService.Register(&req)
	if err != nil {
		log.Printf("Failed to register user: %v", err)
		status := http.StatusInternalServerError
		switch err.Error() {
		case "username 不能为空",
			"username 长度不能超过 50 个字符",
			"username 只能包含字母、数字和下划线",
			"display_name 长度不能超过 100 个字符",
			"email 不能为空",
			"email 长度不能超过 255 个字符",
			"password 不能为空",
			"password 长度必须在 6 到 72 个字符之间":
			status = http.StatusBadRequest
		case "username 已存在", "email 已被使用":
			status = http.StatusConflict
		}
		ctx.JSON(status, gin.H{"code": status, "message": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"code":    http.StatusCreated,
		"message": "用户注册成功",
		"data": gin.H{
			"user": gin.H{
				"id":           user.ID,
				"username":     user.Username,
				"display_name": user.DisplayName,
				"email":        user.Email,
			},
		},
	},
	)
}

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

func Login(ctx *gin.Context) {
	var req models.LoginRequest
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

func RefreshToken(ctx *gin.Context) {
	var req models.RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请求参数无效，需要提供 refresh_token",
		})
		return
	}

	refreshTokenResponse, err := userService.RefreshToken(req.RefreshToken)
	if err != nil {
		log.Printf("Failed to refresh token: %v", err)
		status := http.StatusUnauthorized
		if err.Error() == "refresh_token 不能为空" {
			status = http.StatusBadRequest
		}
		ctx.JSON(status, gin.H{
			"code":    status,
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
	var req models.LogoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请求参数无效，需要提供 refresh_token",
		})
		return
	}
	userID, ok := ctx.Get("userID")
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权访问",
		})
		return
	}
	if err := userService.ValidateRefreshTokenForUser(userID.(uint), req.RefreshToken); err != nil {
		status := http.StatusUnauthorized
		if err.Error() == "refresh_token 不能为空" {
			status = http.StatusBadRequest
		}
		ctx.JSON(status, gin.H{
			"code":    status,
			"message": err.Error(),
		})
		return
	}
	req.RefreshToken = strings.TrimSpace(req.RefreshToken)

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
		status := http.StatusInternalServerError
		switch err.Error() {
		case "用户不存在":
			status = http.StatusNotFound
		case "头像文件不能为空", "头像文件大小超出限制", "头像文件类型无效", "读取头像文件失败":
			status = http.StatusBadRequest
		}
		ctx.JSON(status, gin.H{
			"code":    status,
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

func UpdateDisplayName(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权访问",
		})
		return
	}

	var req models.UpdateDisplayNameRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请求参数无效",
		})
		return
	}

	if err := userService.UpdateDisplayName(userID.(uint), req.DisplayName); err != nil {
		log.Printf("Failed to update display name: %v", err)
		status := http.StatusInternalServerError
		if err.Error() == "display_name 不能为空" || err.Error() == "display_name 长度不能超过 100 个字符" {
			status = http.StatusBadRequest
		} else if err.Error() == "用户不存在" {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{
			"code":    status,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "显示名称更新成功",
	})
}

func UpdatePassword(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权访问",
		})
		return
	}

	var req models.UpdatePasswordRequest
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

package controllers

import (
	"chronote/models"
	"chronote/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

var userService = services.UserService{}

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

func Login(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": http.StatusBadRequest, "message": "请求参数无效"})
		return
	}
	loginResponse, err := userService.Login(user.Email, user.Password)
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

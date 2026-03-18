package controllers

import (
	"chronote/dto"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UploadMedia(ctx *gin.Context) {
	userID, ok := getUserID(ctx)
	if !ok {
		return
	}
	postcardID, err := parseUintParam(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "明信片 ID 无效",
		})
		return
	}
	if _, err := postcardService.EnsureOwner(userID, postcardID); err != nil {
		log.Printf("Failed to verify postcard owner: %v", err)
		status := http.StatusForbidden
		if err.Error() == "明信片不存在" {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{
			"code":    status,
			"message": err.Error(),
		})
		return
	}
	form, err := ctx.MultipartForm()
	if err != nil || form == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请上传媒体文件",
		})
		return
	}
	files := form.File["media"]
	if len(files) == 0 {
		files = form.File["medias"]
	}
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请上传媒体文件",
		})
		return
	}
	mediaType := ctx.PostForm("media_type")
	mediaGroup := ctx.PostForm("media_group")
	medias, err := mediaService.BatchProcessAndUpload(postcardID, files, mediaType, mediaGroup)
	if err != nil {
		log.Printf("Failed to upload media: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "媒体上传成功",
		"data": gin.H{
			"medias": dto.NewMediaResponses(medias),
		},
	})
}

func GetMedias(ctx *gin.Context) {
	userID := getOptionalUserID(ctx)
	postcardID, err := parseUintParam(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "明信片 ID 无效",
		})
		return
	}
	postcard, err := postcardService.GetDetail(userID, postcardID)
	if err != nil {
		log.Printf("Failed to get postcard for media list: %v", err)
		status := http.StatusBadRequest
		if err.Error() == "明信片不存在" {
			status = http.StatusNotFound
		} else if err.Error() == "无权限访问该明信片" {
			status = http.StatusForbidden
		}
		ctx.JSON(status, gin.H{
			"code":    status,
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "获取媒体列表成功",
		"data": gin.H{
			"medias": dto.NewMediaResponses(postcard.Medias),
		},
	})
}

func ReorderMedia(ctx *gin.Context) {
	userID, ok := getUserID(ctx)
	if !ok {
		return
	}
	postcardID, err := parseUintParam(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "明信片 ID 无效",
		})
		return
	}
	if _, err := postcardService.EnsureOwner(userID, postcardID); err != nil {
		log.Printf("Failed to verify postcard owner: %v", err)
		status := http.StatusForbidden
		if err.Error() == "明信片不存在" {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{
			"code":    status,
			"message": err.Error(),
		})
		return
	}
	var req dto.ReorderMediaRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请求参数无效",
		})
		return
	}
	if err := mediaService.Reorder(postcardID, req.MediaIDs); err != nil {
		log.Printf("Failed to reorder media: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "媒体排序更新成功",
	})
}

func DeleteMedia(ctx *gin.Context) {
	userID, ok := getUserID(ctx)
	if !ok {
		return
	}
	postcardID, err := parseUintParam(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "明信片 ID 无效",
		})
		return
	}
	mediaID, err := parseUintParam(ctx.Param("media_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "媒体 ID 无效",
		})
		return
	}
	if _, err := postcardService.EnsureOwner(userID, postcardID); err != nil {
		log.Printf("Failed to verify postcard owner: %v", err)
		status := http.StatusForbidden
		if err.Error() == "明信片不存在" {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{
			"code":    status,
			"message": err.Error(),
		})
		return
	}
	if err := mediaService.Delete(postcardID, mediaID); err != nil {
		log.Printf("Failed to delete media: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "媒体删除成功",
	})
}

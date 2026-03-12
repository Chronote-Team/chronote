package controllers

import (
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"chronote/models"
	"chronote/services"
)

var postcardService = services.PostcardService{}
var mediaService = services.MediaService{}

// CreatePostcard godoc
// @Summary 创建明信片
// @Tags Postcard
// @Accept json
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param title formData string false "标题"
// @Param content formData string false "内容JSON"
// @Param visibility formData string false "可见性"
// @Param media formData file false "媒体文件"
// @Param body body models.CreatePostcardRequest false "创建明信片请求"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /v1/postcards [post]
func CreatePostcard(ctx *gin.Context) {
	userID, ok := getUserID(ctx)
	if !ok {
		return
	}

	var req models.CreatePostcardRequest
	isMultipart := strings.HasPrefix(ctx.ContentType(), "multipart/")
	if isMultipart {
		title := strings.TrimSpace(ctx.PostForm("title"))
		content := strings.TrimSpace(ctx.PostForm("content"))
		if title == "" || content == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "title 和 content 为必填",
			})
			return
		}
		req = models.CreatePostcardRequest{
			Title:      title,
			Content:    json.RawMessage(content),
			Visibility: ctx.PostForm("visibility"),
		}
	} else {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "请求参数无效",
			})
			return
		}
	}
	if !json.Valid(req.Content) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "content 无效",
		})
		return
	}

	var err error
	mediaFiles := make([]*multipart.FileHeader, 0)
	if isMultipart {
		mediaFiles, err = extractMediaFiles(ctx)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    http.StatusBadRequest,
				"message": "请求参数无效",
			})
			return
		}
	}
	postcard, err := postcardService.Create(userID, &req, mediaFiles, ctx.PostForm("media_type"), ctx.PostForm("media_group"))
	if err != nil {
		log.Printf("Failed to create postcard: %v", err)
		status := http.StatusBadRequest
		if err.Error() == "创建明信片失败" {
			status = http.StatusInternalServerError
		}
		ctx.JSON(status, gin.H{
			"code":    status,
			"message": err.Error(),
		})
		return
	}

	detail, err := postcardService.GetDetail(userID, postcard.ID)
	if err != nil {
		log.Printf("Failed to fetch postcard detail: %v", err)
		detail = postcard
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"code":    http.StatusCreated,
		"message": "明信片创建成功",
		"data":    buildPostcardResponse(detail),
	})
}

// GetPostcards godoc
// @Summary 获取明信片列表
// @Tags Postcard
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param visibility query string false "可见性"
// @Param sort_by query string false "排序字段"
// @Param order query string false "排序方向"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /v1/postcards [get]
func GetPostcards(ctx *gin.Context) {
	userID, ok := getUserID(ctx)
	if !ok {
		return
	}
	var query models.PostcardListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请求参数无效",
		})
		return
	}
	postcards, pagination, err := postcardService.List(userID, query)
	if err != nil {
		log.Printf("Failed to list postcards: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}
	responses := make([]models.PostcardResponse, 0, len(postcards))
	for i := range postcards {
		responses = append(responses, buildPostcardResponse(&postcards[i]))
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "获取明信片列表成功",
		"data": gin.H{
			"items":      responses,
			"pagination": pagination,
		},
	})
}

// GetPostcardDetail godoc
// @Summary 获取明信片详情
// @Tags Postcard
// @Produce json
// @Security BearerAuth
// @Param id path int true "明信片ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /v1/postcards/{id} [get]
func GetPostcardDetail(ctx *gin.Context) {
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
	postcard, err := postcardService.GetDetail(userID, postcardID)
	if err != nil {
		log.Printf("Failed to get postcard detail: %v", err)
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
		"message": "获取明信片详情成功",
		"data":    buildPostcardResponse(postcard),
	})
}

// UpdatePostcard godoc
// @Summary 更新明信片
// @Tags Postcard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "明信片ID"
// @Param body body models.UpdatePostcardRequest true "更新明信片请求"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /v1/postcards/{id} [put]
func UpdatePostcard(ctx *gin.Context) {
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
	var req models.UpdatePostcardRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code":    http.StatusBadRequest,
			"message": "请求参数无效",
		})
		return
	}
	if err := postcardService.Update(userID, postcardID, &req); err != nil {
		log.Printf("Failed to update postcard: %v", err)
		status := http.StatusBadRequest
		if err.Error() == "明信片不存在" {
			status = http.StatusNotFound
		} else if err.Error() == "无权限操作该明信片" {
			status = http.StatusForbidden
		} else if err.Error() == "更新明信片失败" {
			status = http.StatusInternalServerError
		}
		ctx.JSON(status, gin.H{
			"code":    status,
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "明信片更新成功",
	})
}

// DeletePostcard godoc
// @Summary 删除明信片
// @Tags Postcard
// @Produce json
// @Security BearerAuth
// @Param id path int true "明信片ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /v1/postcards/{id} [delete]
func DeletePostcard(ctx *gin.Context) {
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
	if err := postcardService.Delete(userID, postcardID); err != nil {
		log.Printf("Failed to delete postcard: %v", err)
		status := http.StatusBadRequest
		if err.Error() == "明信片不存在" {
			status = http.StatusNotFound
		} else if err.Error() == "无权限操作该明信片" {
			status = http.StatusForbidden
		} else if err.Error() == "删除明信片失败" {
			status = http.StatusInternalServerError
		}
		ctx.JSON(status, gin.H{
			"code":    status,
			"message": err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": "明信片删除成功",
	})
}

func buildPostcardResponse(postcard *models.Postcard) models.PostcardResponse {
	var author *models.PostcardAuthorResponse
	if postcard.Author != nil {
		author = &models.PostcardAuthorResponse{
			ID:          postcard.Author.ID,
			Username:    postcard.Author.Username,
			DisplayName: postcard.Author.DisplayName,
			Avatar:      postcard.Author.Avatar,
		}
	}
	return models.PostcardResponse{
		ID:         postcard.ID,
		Title:      postcard.Title,
		Content:    postcard.Content,
		Visibility: postcard.Visibility,
		AuthorID:   postcard.AuthorID,
		Author:     author,
		Medias:     postcard.Medias,
		CreatedAt:  postcard.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  postcard.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func getUserID(ctx *gin.Context) (uint, bool) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权访问",
		})
		return 0, false
	}
	userIDValue, ok := userID.(uint)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code":    http.StatusUnauthorized,
			"message": "未授权访问",
		})
		return 0, false
	}
	return userIDValue, true
}

func parseUintParam(value string) (uint, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func extractMediaFiles(ctx *gin.Context) ([]*multipart.FileHeader, error) {
	form, err := ctx.MultipartForm()
	if err != nil {
		return nil, err
	}
	if form == nil {
		return nil, nil
	}
	files := form.File["media"]
	if len(files) == 0 {
		files = form.File["medias"]
	}
	if len(files) == 0 {
		return nil, nil
	}
	return files, nil
}

package http

import (
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	mediaapp "chronote-refactor/internal/modules/media/app"
	postcardsapp "chronote-refactor/internal/modules/postcards/app"
	"chronote-refactor/internal/shared/errs"
	"chronote-refactor/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	postcards *postcardsapp.Service
	media     *mediaapp.Service
}

func NewHandler(postcards *postcardsapp.Service, media *mediaapp.Service) *Handler {
	return &Handler{postcards: postcards, media: media}
}

func (h *Handler) CreatePostcard(ctx *gin.Context) {
	userID, ok := ctx.Get("userID")
	if !ok {
		response.Write(ctx, http.StatusUnauthorized, "未授权访问", nil)
		return
	}

	var req CreatePostcardRequest
	isMultipart := strings.HasPrefix(ctx.ContentType(), "multipart/")
	if isMultipart {
		title := strings.TrimSpace(ctx.PostForm("title"))
		content := strings.TrimSpace(ctx.PostForm("content"))
		if title == "" || content == "" {
			response.Write(ctx, http.StatusBadRequest, "title 和 content 为必填", nil)
			return
		}
		req = CreatePostcardRequest{
			Title:      title,
			Content:    json.RawMessage(content),
			Visibility: ctx.PostForm("visibility"),
		}
	} else {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			response.Write(ctx, http.StatusBadRequest, "请求参数无效", nil)
			return
		}
	}
	if !json.Valid(req.Content) {
		response.Write(ctx, http.StatusBadRequest, "content 无效", nil)
		return
	}

	postcard, err := h.postcards.Create(userID.(uint), postcardsapp.CreateInput{
		Title:      req.Title,
		Content:    req.Content,
		Visibility: req.Visibility,
	})
	if err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}

	if isMultipart {
		files, err := extractMediaFiles(ctx)
		if err != nil {
			_ = h.postcards.Delete(userID.(uint), postcard.ID)
			response.Write(ctx, http.StatusBadRequest, "请求参数无效", nil)
			return
		}
		if len(files) > 0 {
			if _, err := h.media.UploadBatch(postcard.ID, files, ctx.PostForm("media_type"), ctx.PostForm("media_group")); err != nil {
				_ = h.postcards.Delete(userID.(uint), postcard.ID)
				status, message := errs.MapHTTP(err)
				response.Write(ctx, status, message, nil)
				return
			}
		}
	}

	detail, err := h.postcards.GetDetail(userID.(uint), postcard.ID)
	if err != nil {
		detail = postcard
	}
	response.Write(ctx, http.StatusCreated, "明信片创建成功", newPostcardResponse(detail))
}

func (h *Handler) GetPostcards(ctx *gin.Context) {
	var query PostcardListQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		response.Write(ctx, http.StatusBadRequest, "请求参数无效", nil)
		return
	}

	var userID uint
	if value, ok := ctx.Get("userID"); ok {
		userID = value.(uint)
	}

	postcards, page, err := h.postcards.List(userID, postcardsapp.ListInput{
		Page:       query.Page,
		PageSize:   query.PageSize,
		Visibility: query.Visibility,
		SortBy:     query.SortBy,
		Order:      query.Order,
	})
	if err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}

	items := make([]PostcardResponse, 0, len(postcards))
	for i := range postcards {
		items = append(items, newPostcardResponse(&postcards[i]))
	}
	response.Write(ctx, http.StatusOK, "获取明信片列表成功", gin.H{
		"items":      items,
		"pagination": page,
	})
}

func (h *Handler) GetPostcardDetail(ctx *gin.Context) {
	postcardID, err := parseUintParam(ctx.Param("id"))
	if err != nil {
		response.Write(ctx, http.StatusBadRequest, "明信片 ID 无效", nil)
		return
	}

	var userID uint
	if value, ok := ctx.Get("userID"); ok {
		userID = value.(uint)
	}

	postcard, err := h.postcards.GetDetail(userID, postcardID)
	if err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	response.Write(ctx, http.StatusOK, "获取明信片详情成功", newPostcardResponse(postcard))
}

func (h *Handler) UpdatePostcard(ctx *gin.Context) {
	userID, ok := ctx.Get("userID")
	if !ok {
		response.Write(ctx, http.StatusUnauthorized, "未授权访问", nil)
		return
	}
	postcardID, err := parseUintParam(ctx.Param("id"))
	if err != nil {
		response.Write(ctx, http.StatusBadRequest, "明信片 ID 无效", nil)
		return
	}

	var req UpdatePostcardRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Write(ctx, http.StatusBadRequest, "请求参数无效", nil)
		return
	}

	if err := h.postcards.Update(userID.(uint), postcardID, postcardsapp.UpdateInput{
		Title:      req.Title,
		Content:    req.Content,
		Visibility: req.Visibility,
	}); err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	response.Write(ctx, http.StatusOK, "明信片更新成功", nil)
}

func (h *Handler) DeletePostcard(ctx *gin.Context) {
	userID, ok := ctx.Get("userID")
	if !ok {
		response.Write(ctx, http.StatusUnauthorized, "未授权访问", nil)
		return
	}
	postcardID, err := parseUintParam(ctx.Param("id"))
	if err != nil {
		response.Write(ctx, http.StatusBadRequest, "明信片 ID 无效", nil)
		return
	}

	if err := h.postcards.Delete(userID.(uint), postcardID); err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	response.Write(ctx, http.StatusOK, "明信片删除成功", nil)
}

func parseUintParam(value string) (uint, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	return uint(parsed), err
}

func extractMediaFiles(ctx *gin.Context) ([]*multipart.FileHeader, error) {
	form, err := ctx.MultipartForm()
	if err != nil || form == nil {
		return nil, nil
	}
	files := form.File["media"]
	if len(files) == 0 {
		files = form.File["medias"]
	}
	return files, nil
}

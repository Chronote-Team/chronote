package http

import (
	"net/http"
	"strconv"

	mediaapp "chronote-refactor/internal/modules/media/app"
	postcardsapp "chronote-refactor/internal/modules/postcards/app"
	"chronote-refactor/internal/shared/errs"
	"chronote-refactor/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	media     *mediaapp.Service
	postcards *postcardsapp.Service
}

func NewHandler(media *mediaapp.Service, postcards *postcardsapp.Service) *Handler {
	return &Handler{media: media, postcards: postcards}
}

func (h *Handler) UploadMedia(ctx *gin.Context) {
	userID, ok := ctx.Get("userID")
	if !ok {
		response.Write(ctx, http.StatusUnauthorized, "未授权访问", nil)
		return
	}
	postcardID, err := parseUint(ctx.Param("id"))
	if err != nil {
		response.Write(ctx, http.StatusBadRequest, "明信片 ID 无效", nil)
		return
	}
	if _, err := h.postcards.EnsureOwner(userID.(uint), postcardID); err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	form, err := ctx.MultipartForm()
	if err != nil || form == nil {
		response.Write(ctx, http.StatusBadRequest, "请上传媒体文件", nil)
		return
	}
	files := form.File["media"]
	if len(files) == 0 {
		files = form.File["medias"]
	}
	if len(files) == 0 {
		response.Write(ctx, http.StatusBadRequest, "请上传媒体文件", nil)
		return
	}

	medias, err := h.media.UploadBatch(postcardID, files, ctx.PostForm("media_type"), ctx.PostForm("media_group"))
	if err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	response.Write(ctx, http.StatusOK, "媒体上传成功", gin.H{"medias": newMediaResponses(medias)})
}

func (h *Handler) GetMedias(ctx *gin.Context) {
	postcardID, err := parseUint(ctx.Param("id"))
	if err != nil {
		response.Write(ctx, http.StatusBadRequest, "明信片 ID 无效", nil)
		return
	}
	var userID uint
	if value, ok := ctx.Get("userID"); ok {
		userID = value.(uint)
	}
	if _, err := h.postcards.GetDetail(userID, postcardID); err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	medias, err := h.media.List(postcardID)
	if err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	response.Write(ctx, http.StatusOK, "获取媒体列表成功", gin.H{"medias": newMediaResponses(medias)})
}

func (h *Handler) ReorderMedia(ctx *gin.Context) {
	userID, ok := ctx.Get("userID")
	if !ok {
		response.Write(ctx, http.StatusUnauthorized, "未授权访问", nil)
		return
	}
	postcardID, err := parseUint(ctx.Param("id"))
	if err != nil {
		response.Write(ctx, http.StatusBadRequest, "明信片 ID 无效", nil)
		return
	}
	if _, err := h.postcards.EnsureOwner(userID.(uint), postcardID); err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	var req ReorderMediaRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Write(ctx, http.StatusBadRequest, "请求参数无效", nil)
		return
	}
	if err := h.media.Reorder(postcardID, req.MediaIDs); err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	response.Write(ctx, http.StatusOK, "媒体排序更新成功", nil)
}

func (h *Handler) DeleteMedia(ctx *gin.Context) {
	userID, ok := ctx.Get("userID")
	if !ok {
		response.Write(ctx, http.StatusUnauthorized, "未授权访问", nil)
		return
	}
	postcardID, err := parseUint(ctx.Param("id"))
	if err != nil {
		response.Write(ctx, http.StatusBadRequest, "明信片 ID 无效", nil)
		return
	}
	mediaID, err := parseUint(ctx.Param("media_id"))
	if err != nil {
		response.Write(ctx, http.StatusBadRequest, "媒体 ID 无效", nil)
		return
	}
	if _, err := h.postcards.EnsureOwner(userID.(uint), postcardID); err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	if err := h.media.Delete(postcardID, mediaID); err != nil {
		status, message := errs.MapHTTP(err)
		response.Write(ctx, status, message, nil)
		return
	}
	response.Write(ctx, http.StatusOK, "媒体删除成功", nil)
}

func parseUint(value string) (uint, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	return uint(parsed), err
}

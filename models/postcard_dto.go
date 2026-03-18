package models

import (
	"encoding/json"

	"gorm.io/datatypes"
)

type CreatePostcardRequest struct {
	Title      string          `json:"title" binding:"required,max=200"`
	Content    json.RawMessage `json:"content" binding:"required"`
	Visibility string          `json:"visibility" binding:"omitempty,oneof=public private"`
}

type UpdatePostcardRequest struct {
	Title      *string          `json:"title"`
	Content    *json.RawMessage `json:"content"`
	Visibility *string          `json:"visibility"`
}

type PostcardListQuery struct {
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=50"`
	Visibility string `form:"visibility" binding:"omitempty,oneof=public private"`
	SortBy     string `form:"sort_by"`
	Order      string `form:"order"`
}

type Pagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
}

type ReorderMediaRequest struct {
	MediaIDs []uint `json:"media_ids" binding:"required,min=1,max=100"`
}

type PostcardAuthorResponse struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar,omitempty"`
}

type PostcardResponse struct {
	ID         uint                    `json:"id"`
	Title      string                  `json:"title"`
	Content    datatypes.JSON          `json:"content"`
	Visibility string                  `json:"visibility"`
	AuthorID   uint                    `json:"author_id"`
	Author     *PostcardAuthorResponse `json:"author,omitempty"`
	Medias     []PostcardMedia         `json:"medias,omitempty"`
	CreatedAt  string                  `json:"created_at"`
	UpdatedAt  string                  `json:"updated_at"`
}

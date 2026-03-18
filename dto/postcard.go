package dto

import (
	"encoding/json"

	"chronote/models"
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

type PostcardAuthorResponse struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Avatar      string `json:"avatar,omitempty"`
}

type PostcardResponse struct {
	ID         uint                    `json:"id"`
	Title      string                  `json:"title"`
	Content    json.RawMessage         `json:"content"`
	Visibility string                  `json:"visibility"`
	AuthorID   uint                    `json:"author_id"`
	Author     *PostcardAuthorResponse `json:"author,omitempty"`
	Medias     []MediaResponse         `json:"medias,omitempty"`
	CreatedAt  string                  `json:"created_at"`
	UpdatedAt  string                  `json:"updated_at"`
}

func NewPostcardResponse(postcard *models.Postcard) PostcardResponse {
	var author *PostcardAuthorResponse
	if postcard.Author != nil {
		author = &PostcardAuthorResponse{
			ID:          postcard.Author.ID,
			Username:    postcard.Author.Username,
			DisplayName: postcard.Author.DisplayName,
			Avatar:      postcard.Author.Avatar,
		}
	}

	return PostcardResponse{
		ID:         postcard.ID,
		Title:      postcard.Title,
		Content:    json.RawMessage(append([]byte(nil), postcard.Content...)),
		Visibility: postcard.Visibility,
		AuthorID:   postcard.AuthorID,
		Author:     author,
		Medias:     NewMediaResponses(postcard.Medias),
		CreatedAt:  postcard.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  postcard.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

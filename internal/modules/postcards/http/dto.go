package http

import (
	"encoding/json"

	mediadomain "chronote-refactor/internal/modules/media/domain"
	postcardsdomain "chronote-refactor/internal/modules/postcards/domain"
)

type CreatePostcardRequest struct {
	Title      string          `json:"title" binding:"required,max=200"`
	Content    json.RawMessage `json:"content" binding:"required"`
	Visibility string          `json:"visibility"`
}

type UpdatePostcardRequest struct {
	Title      *string          `json:"title"`
	Content    *json.RawMessage `json:"content"`
	Visibility *string          `json:"visibility"`
}

type PostcardListQuery struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	Visibility string `form:"visibility"`
	SortBy     string `form:"sort_by"`
	Order      string `form:"order"`
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
	Medias     []PostcardMediaResponse `json:"medias,omitempty"`
	CreatedAt  string                  `json:"created_at"`
	UpdatedAt  string                  `json:"updated_at"`
}

type PostcardMediaResponse struct {
	ID             uint   `json:"id"`
	PostcardID     uint   `json:"postcard_id"`
	Type           string `json:"type"`
	URL            string `json:"url"`
	ThumbnailURL   string `json:"thumbnail_url,omitempty"`
	OriginalWidth  int    `json:"original_width,omitempty"`
	OriginalHeight int    `json:"original_height,omitempty"`
	Duration       int    `json:"duration,omitempty"`
	FileSize       int64  `json:"file_size"`
	Position       int    `json:"position"`
	Group          string `json:"group"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

func newPostcardResponse(postcard *postcardsdomain.Postcard) PostcardResponse {
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
		Medias:     postcardMediaResponses(postcard.Medias),
		CreatedAt:  postcard.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  postcard.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func postcardMediaResponses(medias []mediadomain.Media) []PostcardMediaResponse {
	if len(medias) == 0 {
		return nil
	}
	responses := make([]PostcardMediaResponse, 0, len(medias))
	for i := range medias {
		responses = append(responses, PostcardMediaResponse{
			ID:             medias[i].ID,
			PostcardID:     medias[i].PostcardID,
			Type:           medias[i].Type,
			URL:            medias[i].URL,
			ThumbnailURL:   medias[i].ThumbnailURL,
			OriginalWidth:  medias[i].OriginalWidth,
			OriginalHeight: medias[i].OriginalHeight,
			Duration:       medias[i].Duration,
			FileSize:       medias[i].FileSize,
			Position:       medias[i].Position,
			Group:          medias[i].Group,
			CreatedAt:      medias[i].CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:      medias[i].UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return responses
}

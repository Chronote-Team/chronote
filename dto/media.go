package dto

import "chronote/models"

type ReorderMediaRequest struct {
	MediaIDs []uint `json:"media_ids" binding:"required,min=1,max=100"`
}

type MediaResponse struct {
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

func NewMediaResponse(media *models.PostcardMedia) MediaResponse {
	return MediaResponse{
		ID:             media.ID,
		PostcardID:     media.PostcardID,
		Type:           media.MediaType,
		URL:            media.URL,
		ThumbnailURL:   media.ThumbnailURL,
		OriginalWidth:  media.OriginalWidth,
		OriginalHeight: media.OriginalHeight,
		Duration:       media.Duration,
		FileSize:       media.FileSize,
		Position:       media.Position,
		Group:          media.MediaGroup,
		CreatedAt:      media.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:      media.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

func NewMediaResponses(medias []models.PostcardMedia) []MediaResponse {
	if len(medias) == 0 {
		return nil
	}

	responses := make([]MediaResponse, 0, len(medias))
	for i := range medias {
		responses = append(responses, NewMediaResponse(&medias[i]))
	}
	return responses
}

package dto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"chronote/models"
)

func TestNewMediaResponse(t *testing.T) {
	now := time.Date(2026, 3, 19, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	media := &models.PostcardMedia{
		BaseModel: models.BaseModel{
			ID:        3,
			CreatedAt: now,
			UpdatedAt: now,
		},
		PostcardID:      12,
		MediaType:       "image",
		OSSKey:          "postcards/internal.jpg",
		URL:             "https://cdn.example.com/postcards/12/photo.jpg",
		ThumbnailURL:    "https://cdn.example.com/postcards/12/thumb.jpg",
		ThumbnailOSSKey: "postcards/internal-thumb.jpg",
		OriginalWidth:   1920,
		OriginalHeight:  1080,
		FileSize:        2048,
		Position:        1,
		MediaGroup:      "gallery",
	}

	response := NewMediaResponse(media)

	if response.ID != 3 || response.PostcardID != 12 {
		t.Fatalf("expected media identifiers to be mapped, got %+v", response)
	}
	if response.Type != "image" || response.Group != "gallery" {
		t.Fatalf("expected media type/group to be mapped, got %+v", response)
	}
	if response.URL != media.URL || response.ThumbnailURL != media.ThumbnailURL {
		t.Fatalf("expected media urls to be mapped, got %+v", response)
	}
	if response.CreatedAt != "2026-03-19 10:00:00" || response.UpdatedAt != "2026-03-19 10:00:00" {
		t.Fatalf("expected timestamps to be formatted, got %+v", response)
	}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	text := string(payload)
	if strings.Contains(text, "oss_key") || strings.Contains(text, "thumbnail_oss_key") {
		t.Fatalf("response leaked internal storage fields: %s", text)
	}
}

func TestNewMediaResponsesReturnsNilForEmptyInput(t *testing.T) {
	if responses := NewMediaResponses(nil); responses != nil {
		t.Fatalf("expected nil for nil input, got %#v", responses)
	}
	if responses := NewMediaResponses([]models.PostcardMedia{}); responses != nil {
		t.Fatalf("expected nil for empty input, got %#v", responses)
	}
}

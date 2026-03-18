package dto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"chronote/models"
	"gorm.io/datatypes"
)

func TestNewPostcardResponseMapsFieldsAndCopiesContent(t *testing.T) {
	now := time.Date(2026, 3, 19, 11, 15, 0, 0, time.FixedZone("CST", 8*3600))
	postcard := &models.Postcard{
		BaseModel: models.BaseModel{
			ID:        5,
			CreatedAt: now,
			UpdatedAt: now,
		},
		Title:      "Spring Note",
		Content:    datatypes.JSON([]byte(`{"blocks":[{"type":"text","value":"hello"}]}`)),
		Visibility: "public",
		AuthorID:   2,
		Author: &models.User{
			BaseModel:   models.BaseModel{ID: 2},
			Username:    "writer",
			DisplayName: "Writer",
			Avatar:      "https://cdn.example.com/avatar.png",
		},
		Medias: []models.PostcardMedia{
			{
				BaseModel:  models.BaseModel{ID: 8, CreatedAt: now, UpdatedAt: now},
				PostcardID: 5,
				MediaType:  "image",
				URL:        "https://cdn.example.com/postcards/5/a.jpg",
				OSSKey:     "internal/a.jpg",
				FileSize:   1024,
				Position:   1,
				MediaGroup: "gallery",
			},
		},
	}

	response := NewPostcardResponse(postcard)

	if response.ID != 5 || response.AuthorID != 2 {
		t.Fatalf("expected identifiers to be mapped, got %+v", response)
	}
	if response.Author == nil || response.Author.Username != "writer" {
		t.Fatalf("expected author dto to be present, got %+v", response.Author)
	}
	if response.CreatedAt != "2026-03-19 11:15:00" || response.UpdatedAt != "2026-03-19 11:15:00" {
		t.Fatalf("expected formatted timestamps, got %+v", response)
	}
	if len(response.Medias) != 1 {
		t.Fatalf("expected one mapped media, got %+v", response.Medias)
	}

	postcard.Content[0] = '['
	if string(response.Content) != `{"blocks":[{"type":"text","value":"hello"}]}` {
		t.Fatalf("expected response content to be copied, got %s", string(response.Content))
	}

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	text := string(payload)
	if strings.Contains(text, "oss_key") || strings.Contains(text, "thumbnail_oss_key") {
		t.Fatalf("response leaked internal media fields: %s", text)
	}
}

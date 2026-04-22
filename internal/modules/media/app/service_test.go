package app

import (
	"testing"

	mediadomain "chronote-refactor/internal/modules/media/domain"
)

func TestValidateGroupCompatibility(t *testing.T) {
	tests := []struct {
		name      string
		mediaType string
		group     string
		wantError string
	}{
		{name: "header allows image", mediaType: "image", group: "header"},
		{name: "header rejects audio", mediaType: "audio", group: "header", wantError: "header 分组仅支持图片"},
		{name: "bgm rejects image", mediaType: "image", group: "bgm", wantError: "bgm 分组仅支持音频"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGroupCompatibility(tt.mediaType, tt.group)
			if tt.wantError != "" {
				if err == nil || err.Error() != tt.wantError {
					t.Fatalf("expected error %q, got %v", tt.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected success, got %v", err)
			}
		})
	}
}

func TestReorderRequiresCompleteMediaSet(t *testing.T) {
	mediaRepo := newMemoryRepository()
	if _, err := mediaRepo.Create(&mediadomain.Media{PostcardID: 9, Type: "image", Group: "gallery", URL: "https://example.com/1.jpg", StorageKey: "1.jpg", FileSize: 1, Position: 1}); err != nil {
		t.Fatalf("seed first media: %v", err)
	}
	if _, err := mediaRepo.Create(&mediadomain.Media{PostcardID: 9, Type: "image", Group: "gallery", URL: "https://example.com/2.jpg", StorageKey: "2.jpg", FileSize: 1, Position: 2}); err != nil {
		t.Fatalf("seed second media: %v", err)
	}

	svc := NewService(mediaRepo, nil, nil)
	if err := svc.Reorder(9, []uint{1}); err == nil || err.Error() != "必须传入全部媒体ID进行排序" {
		t.Fatalf("expected complete-order validation error, got %v", err)
	}
}

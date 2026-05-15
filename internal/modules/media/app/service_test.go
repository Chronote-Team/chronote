package app

import (
	"context"
	"testing"

	mediadomain "chronote-refactor/internal/modules/media/domain"
	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
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

func TestMediaReorderAndDeleteEnqueuePostcardAnalysis(t *testing.T) {
	mediaRepo := newMemoryRepository()
	first, err := mediaRepo.Create(&mediadomain.Media{PostcardID: 9, Type: "image", Group: "gallery", URL: "https://example.com/1.jpg", StorageKey: "1.jpg", FileSize: 1, Position: 1})
	if err != nil {
		t.Fatalf("seed first media: %v", err)
	}
	second, err := mediaRepo.Create(&mediadomain.Media{PostcardID: 9, Type: "image", Group: "gallery", URL: "https://example.com/2.jpg", StorageKey: "2.jpg", FileSize: 1, Position: 2})
	if err != nil {
		t.Fatalf("seed second media: %v", err)
	}

	svc := NewService(mediaRepo, nil, nil)
	enqueuer := &recordingAnalysisEnqueuer{}
	svc.SetAnalysisEnqueuer(enqueuer)

	if err := svc.Reorder(9, []uint{second.ID, first.ID}); err != nil {
		t.Fatalf("Reorder returned error: %v", err)
	}
	if err := svc.Delete(9, first.ID); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if len(enqueuer.inputs) != 2 {
		t.Fatalf("expected two media-change enqueues, got %#v", enqueuer.inputs)
	}
	for _, input := range enqueuer.inputs {
		if input.Reason != postcardaiapp.EnqueueReasonMediaChange || input.PostcardID != 9 {
			t.Fatalf("unexpected enqueue input %#v", input)
		}
	}
}

type recordingAnalysisEnqueuer struct {
	inputs []postcardaiapp.EnqueueInput
}

func (r *recordingAnalysisEnqueuer) EnqueuePostcardAnalysis(ctx context.Context, input postcardaiapp.EnqueueInput) (*postcardaiapp.EnqueueResult, error) {
	r.inputs = append(r.inputs, input)
	return &postcardaiapp.EnqueueResult{}, nil
}

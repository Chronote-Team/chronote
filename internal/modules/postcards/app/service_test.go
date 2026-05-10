package app

import (
	"encoding/json"
	"strings"
	"testing"

	postcardsdomain "chronote-refactor/internal/modules/postcards/domain"
	"chronote-refactor/internal/shared/errs"
)

func TestValidatePostcardTitle(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantError string
	}{
		{name: "trimmed title", input: "  Hello  ", want: "Hello"},
		{name: "empty title", input: "   ", wantError: "title 不能为空"},
		{name: "too long", input: strings.Repeat("a", 201), wantError: "title 长度不能超过 200 个字符"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateTitle(tt.input)
			if tt.wantError != "" {
				if err == nil || err.Error() != tt.wantError {
					t.Fatalf("expected error %q, got %v", tt.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected success, got error %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestValidatePostcardContent(t *testing.T) {
	tests := []struct {
		name      string
		input     json.RawMessage
		want      string
		wantError string
	}{
		{name: "empty content", input: nil, wantError: "content 不能为空"},
		{name: "invalid json", input: json.RawMessage(`{"broken"`), wantError: "content 无效"},
		{name: "scalar json", input: json.RawMessage(`"text"`), wantError: "content 必须是 JSON 对象或数组"},
		{name: "object json", input: json.RawMessage(`{"text":"hello"}`), want: `{"text":"hello"}`},
		{name: "array json", input: json.RawMessage(`[{"type":"text"}]`), want: `[{"type":"text"}]`},
		{name: "too large", input: json.RawMessage(`{"text":"` + strings.Repeat("a", maxContentBytes) + `"}`), wantError: "content 长度不能超过 65536 字节"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateContent(tt.input)
			if tt.wantError != "" {
				if err == nil || err.Error() != tt.wantError {
					t.Fatalf("expected error %q, got %v", tt.wantError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected success, got error %v", err)
			}
			if string(got) != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, string(got))
			}
		})
	}
}

func TestCanAccessPostcard(t *testing.T) {
	tests := []struct {
		name       string
		userID     uint
		authorID   uint
		visibility string
		want       bool
	}{
		{name: "owner can access", userID: 1, authorID: 1, visibility: "private", want: true},
		{name: "public postcard visible", userID: 2, authorID: 1, visibility: "public", want: true},
		{name: "private postcard hidden", userID: 2, authorID: 1, visibility: "private", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canAccess(tt.userID, tt.authorID, tt.visibility); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestGetRandomReturnsOnlyPublicPostcardsForAnonymousUser(t *testing.T) {
	service := NewService(nil, nil, nil)
	createPostcardForTest(t, service, 1, "Private Card", "private")
	createPostcardForTest(t, service, 2, "Public Card", "public")

	postcard, err := service.GetRandom(0)
	if err != nil {
		t.Fatalf("expected random postcard, got error %v", err)
	}
	if postcard.Visibility != "public" {
		t.Fatalf("expected public postcard for anonymous user, got %q", postcard.Visibility)
	}
}

func TestGetRandomIncludesOwnedPrivatePostcardsForAuthenticatedUser(t *testing.T) {
	service := NewService(nil, nil, nil)
	createPostcardForTest(t, service, 1, "Other Private Card", "private")
	owned := createPostcardForTest(t, service, 2, "Owned Private Card", "private")

	postcard, err := service.GetRandom(2)
	if err != nil {
		t.Fatalf("expected random postcard, got error %v", err)
	}
	if postcard.ID != owned.ID {
		t.Fatalf("expected owned private postcard %d, got %d", owned.ID, postcard.ID)
	}
}

func TestGetRandomReturnsNotFoundWhenNoPostcardsAreAccessible(t *testing.T) {
	service := NewService(nil, nil, nil)
	createPostcardForTest(t, service, 1, "Other Private Card", "private")

	postcard, err := service.GetRandom(0)
	if postcard != nil {
		t.Fatalf("expected nil postcard, got %#v", postcard)
	}
	appErr, ok := err.(*errs.AppError)
	if !ok {
		t.Fatalf("expected AppError, got %T %v", err, err)
	}
	if appErr.Status != 404 || appErr.Message != "明信片不存在" {
		t.Fatalf("expected not found postcard error, got status=%d message=%q", appErr.Status, appErr.Message)
	}
}

func createPostcardForTest(t *testing.T, service *Service, userID uint, title, visibility string) *postcardsdomain.Postcard {
	t.Helper()

	postcard, err := service.Create(userID, CreateInput{
		Title:      title,
		Content:    json.RawMessage(`{"blocks":[{"type":"text","value":"hello"}]}`),
		Visibility: visibility,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	return postcard
}

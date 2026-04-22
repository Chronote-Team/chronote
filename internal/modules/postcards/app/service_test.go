package app

import (
	"encoding/json"
	"strings"
	"testing"
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

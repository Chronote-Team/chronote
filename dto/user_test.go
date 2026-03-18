package dto

import (
	"testing"
	"time"

	"chronote/models"
)

func TestNewRegisterUserResponse(t *testing.T) {
	user := &models.User{
		BaseModel: models.BaseModel{ID: 7},
		Username:  "alice",
		Email:     "alice@example.com",
		Password:  "hashed-password",
	}

	response := NewRegisterUserResponse(user)

	if response.ID != 7 {
		t.Fatalf("expected id 7, got %d", response.ID)
	}
	if response.Username != "alice" {
		t.Fatalf("expected username alice, got %q", response.Username)
	}
	if response.Email != "alice@example.com" {
		t.Fatalf("expected email alice@example.com, got %q", response.Email)
	}
}

func TestNewUserInfoResponse(t *testing.T) {
	createdAt := time.Date(2026, 3, 19, 9, 30, 45, 0, time.FixedZone("CST", 8*3600))
	user := &models.User{
		BaseModel: models.BaseModel{
			ID:        9,
			CreatedAt: createdAt,
		},
		Username:    "bob",
		DisplayName: "Bob",
		Email:       "bob@example.com",
		Avatar:      "https://cdn.example.com/avatar.jpg",
		Password:    "hashed-password",
	}

	response := NewUserInfoResponse(user)

	if response.ID != 9 {
		t.Fatalf("expected id 9, got %d", response.ID)
	}
	if response.Username != "bob" {
		t.Fatalf("expected username bob, got %q", response.Username)
	}
	if response.DisplayName != "Bob" {
		t.Fatalf("expected display name Bob, got %q", response.DisplayName)
	}
	if response.Email != "bob@example.com" {
		t.Fatalf("expected email bob@example.com, got %q", response.Email)
	}
	if response.Avatar != "https://cdn.example.com/avatar.jpg" {
		t.Fatalf("expected avatar url to be mapped")
	}
	if response.CreatedAt != "2026-03-19 09:30:45" {
		t.Fatalf("expected formatted created_at, got %q", response.CreatedAt)
	}
}

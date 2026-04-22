package app

import "testing"

func TestRegisterNormalizesEmailAndDefaultsDisplayName(t *testing.T) {
	service := NewService(nil, nil)

	user, err := service.Register(RegisterInput{
		Username: "tester",
		Email:    "TEST@EXAMPLE.COM",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	if user.Email != "test@example.com" {
		t.Fatalf("expected lowercased email, got %q", user.Email)
	}
	if user.DisplayName != "tester" {
		t.Fatalf("expected default display name, got %q", user.DisplayName)
	}
}

package app

import "testing"

func TestValidateAccessTokenRejectsRefreshToken(t *testing.T) {
	service := NewService(nil, nil, nil)

	if err := service.ValidateTokenType("refresh"); err == nil {
		t.Fatalf("expected refresh token to be rejected for access validation")
	}
}

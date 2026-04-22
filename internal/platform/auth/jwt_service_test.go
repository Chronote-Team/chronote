package auth

import "testing"

func TestJWTServiceGeneratesAccessAndRefreshTokens(t *testing.T) {
	service := NewJWTService("secret-key", 7200, 1814400)

	accessToken, refreshToken, err := service.GenerateTokenPair(7, "tester")
	if err != nil {
		t.Fatalf("GenerateTokenPair returned error: %v", err)
	}
	if accessToken == "" || refreshToken == "" {
		t.Fatalf("expected both tokens to be generated")
	}

	accessClaims, err := service.ParseToken(accessToken)
	if err != nil {
		t.Fatalf("ParseToken(access) returned error: %v", err)
	}
	if accessClaims.TokenType != "access" || accessClaims.UserID != 7 || accessClaims.Name != "tester" {
		t.Fatalf("unexpected access claims: %+v", accessClaims)
	}

	refreshClaims, err := service.ParseToken(refreshToken)
	if err != nil {
		t.Fatalf("ParseToken(refresh) returned error: %v", err)
	}
	if refreshClaims.TokenType != "refresh" || refreshClaims.UserID != 7 {
		t.Fatalf("unexpected refresh claims: %+v", refreshClaims)
	}
}

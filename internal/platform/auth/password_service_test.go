package auth

import "testing"

func TestPasswordServiceHashesAndVerifiesPassword(t *testing.T) {
	service := PasswordService{}

	hashed, err := service.Hash("password123")
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}
	if hashed == "" || hashed == "password123" {
		t.Fatalf("expected hashed password, got %q", hashed)
	}

	ok, err := service.Verify("password123", hashed)
	if err != nil {
		t.Fatalf("Verify returned error: %v", err)
	}
	if !ok {
		t.Fatalf("expected password to verify")
	}
}

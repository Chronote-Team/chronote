package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type PasswordService struct{}

func (PasswordService) Hash(plainPassword string) (string, error) {
	saltBytes := make([]byte, 12)
	if _, err := rand.Read(saltBytes); err != nil {
		return "", err
	}
	salt := "salt" + hex.EncodeToString(saltBytes)
	passwordWithSalt := plainPassword + salt
	hasher := sha256.New()
	hasher.Write([]byte(passwordWithSalt))
	hashedPassword := hex.EncodeToString(hasher.Sum(nil))
	return fmt.Sprintf("sha256:%s:%s", hashedPassword, salt), nil
}

func (PasswordService) Verify(plainPassword, storedPassword string) (bool, error) {
	parts := strings.Split(storedPassword, ":")
	if len(parts) != 3 {
		return false, fmt.Errorf("Invalid Stored Password!")
	}
	hashedPassword := parts[1]
	salt := parts[2]
	passwordWithSalt := plainPassword + salt
	hasher := sha256.New()
	hasher.Write([]byte(passwordWithSalt))
	inputPassword := hex.EncodeToString(hasher.Sum(nil))
	return inputPassword == hashedPassword, nil
}

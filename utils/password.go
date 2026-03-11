package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func EncryptPassword(plainPassword string) (string, error) {
	saltbytes := make([]byte, 12)
	if _, err := rand.Read(saltbytes); err != nil {
		return "", err
	}
	salt := "salt" + hex.EncodeToString(saltbytes)
	passwordWithSalt := plainPassword + salt
	hasher := sha256.New()
	hasher.Write([]byte(passwordWithSalt))
	hashedPassword := hex.EncodeToString(hasher.Sum(nil))
	method := "sha256"
	final := fmt.Sprintf("%s:%s:%s", method, hashedPassword, salt)
	return final, nil
}

func VerifyPassword(plainPassword, storedPassword string) (bool, error) {
	parts := strings.Split(storedPassword, ":")
	if len(parts) != 3 {
		return false, fmt.Errorf("Invalid Stored Password!")
	}
	// method := parts[0]
	hashedPassword := parts[1]
	salt := parts[2]
	passwordWithSalt := plainPassword + salt
	hasher := sha256.New()
	hasher.Write([]byte(passwordWithSalt))
	inputPassword := hex.EncodeToString(hasher.Sum(nil))
	return inputPassword == hashedPassword, nil
}

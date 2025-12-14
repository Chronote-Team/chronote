package utils

import (
	"chronote/config"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Customnize a User Claim Struct, Which can add more permissions for users
type CustomUserClaim struct {
	UserID    uint   `json:"user_id"`
	Name      string `json:"username"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// Helper functions to get JWT configuration values
func getSigningKey() []byte {
	return []byte(config.AppConfig.JWT.MySigningKey)
}

func getAccessTokenExpire() time.Duration {
	return time.Duration(config.AppConfig.JWT.AccessTokenExpire) * time.Second
}

func getRefreshTokenExpire() time.Duration {
	return time.Duration(config.AppConfig.JWT.RefreshTokenExpire) * time.Second
}

// Generate JWT For users, containing info of app
func GenerateToken(userID uint, Username string) (accessToken, refreshToken string, err error) {
	signingKey := getSigningKey()
	accessTokenExpire := getAccessTokenExpire()
	refreshTokenExpire := getRefreshTokenExpire()

	// Generate Access Token
	accessClaims := CustomUserClaim{
		UserID:    userID,
		Name:      Username,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExpire)),
			Issuer:    "Chronote",
		},
	}
	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(signingKey)
	if err != nil {
		return "", "", err
	}

	// Generate Refresh Token
	refreshClaims := CustomUserClaim{
		UserID:    userID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenExpire)),
			Issuer:    "Chronote",
		},
	}
	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(signingKey)
	if err != nil {
		return "", "", err
	}

	// Return All values
	return accessToken, refreshToken, nil
}

// middlewares should check the token type, it should be `access` when get protected resources
func Parsetoken(tokenString string) (*CustomUserClaim, error) {
	signingKey := getSigningKey()

	// Parse Token
	token, err := jwt.ParseWithClaims(tokenString, &CustomUserClaim{}, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	// Verify Token and Get Claims
	if claims, ok := token.Claims.(*CustomUserClaim); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

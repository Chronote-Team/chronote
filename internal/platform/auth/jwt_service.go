package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type CustomUserClaim struct {
	UserID    uint   `json:"user_id"`
	Name      string `json:"username"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type JWTService struct {
	signingKey       []byte
	accessTokenTTL   time.Duration
	refreshTokenTTL  time.Duration
	accessTokenSecs  int64
	refreshTokenSecs int64
}

func NewJWTService(signingKey string, accessTokenSecs, refreshTokenSecs int64) *JWTService {
	return &JWTService{
		signingKey:       []byte(signingKey),
		accessTokenTTL:   time.Duration(accessTokenSecs) * time.Second,
		refreshTokenTTL:  time.Duration(refreshTokenSecs) * time.Second,
		accessTokenSecs:  accessTokenSecs,
		refreshTokenSecs: refreshTokenSecs,
	}
}

func (s *JWTService) GenerateTokenPair(userID uint, username string) (string, string, error) {
	accessClaims := CustomUserClaim{
		UserID:    userID,
		Name:      username,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.accessTokenTTL)),
			Issuer:    "Chronote",
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(s.signingKey)
	if err != nil {
		return "", "", err
	}

	refreshClaims := CustomUserClaim{
		UserID:    userID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.refreshTokenTTL)),
			Issuer:    "Chronote",
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(s.signingKey)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (s *JWTService) ParseToken(tokenString string) (*CustomUserClaim, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomUserClaim{}, func(token *jwt.Token) (interface{}, error) {
		return s.signingKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*CustomUserClaim)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *JWTService) AccessExpirySeconds() int64 {
	return s.accessTokenSecs
}

package http

import (
	"net/http"
	"strings"

	authapp "chronote-refactor/internal/modules/auth/app"
	"chronote-refactor/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type TokenParser interface {
	ParseToken(token string) (*authapp.TokenClaims, error)
}

type Middleware struct {
	tokens    TokenParser
	blacklist authapp.TokenBlacklist
}

func NewMiddleware(tokens TokenParser, blacklist authapp.TokenBlacklist) *Middleware {
	return &Middleware{tokens: tokens, blacklist: blacklist}
}

func (m *Middleware) RequireAccessToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			response.Write(ctx, http.StatusUnauthorized, "Without Token, and you're unauthorized!", nil)
			ctx.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Write(ctx, http.StatusUnauthorized, "Token format is invalid!", nil)
			ctx.Abort()
			return
		}

		tokenString := parts[1]
		blacklisted, err := m.blacklist.IsBlacklisted(tokenString)
		if err != nil {
			response.Write(ctx, http.StatusInternalServerError, "Failed to validate token!", nil)
			ctx.Abort()
			return
		}
		if blacklisted {
			response.Write(ctx, http.StatusUnauthorized, "Token has been revoked!", nil)
			ctx.Abort()
			return
		}

		claims, err := m.tokens.ParseToken(tokenString)
		if err != nil {
			response.Write(ctx, http.StatusUnauthorized, "Token is invalid or outdated!", nil)
			ctx.Abort()
			return
		}
		if claims.TokenType != "access" {
			response.Write(ctx, http.StatusUnauthorized, "Need Use Aceess Token to Authorize!", nil)
			ctx.Abort()
			return
		}

		ctx.Set("userID", claims.UserID)
		ctx.Set("username", claims.Name)
		ctx.Next()
	}
}

func (m *Middleware) OptionalAccessToken() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			ctx.Next()
			return
		}

		tokenString := parts[1]
		blacklisted, err := m.blacklist.IsBlacklisted(tokenString)
		if err != nil || blacklisted {
			ctx.Next()
			return
		}

		claims, err := m.tokens.ParseToken(tokenString)
		if err != nil || claims.TokenType != "access" {
			ctx.Next()
			return
		}

		ctx.Set("userID", claims.UserID)
		ctx.Set("username", claims.Name)
		ctx.Next()
	}
}

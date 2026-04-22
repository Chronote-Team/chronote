package http

import (
	authhttp "chronote-refactor/internal/modules/auth/http"
	healthhttp "chronote-refactor/internal/modules/health/http"
	mediahttp "chronote-refactor/internal/modules/media/http"
	postcardshttp "chronote-refactor/internal/modules/postcards/http"
	usershttp "chronote-refactor/internal/modules/users/http"

	"github.com/gin-gonic/gin"
)

type RouterDependencies struct {
	HealthHandler       *healthhttp.Handler
	UserHandler         *usershttp.Handler
	AuthHandler         *authhttp.Handler
	PostcardHandler     *postcardshttp.Handler
	MediaHandler        *mediahttp.Handler
	RequireAccessToken  gin.HandlerFunc
	OptionalAccessToken gin.HandlerFunc
}

func NewRouter(deps RouterDependencies) *gin.Engine {
	router := gin.Default()

	healthhttp.RegisterRoutes(router, deps.HealthHandler)

	publicUser := router.Group("/user")
	usershttp.RegisterPublicRoutes(publicUser, deps.UserHandler)
	authhttp.RegisterPublicRoutes(publicUser, deps.AuthHandler)

	protectedUser := router.Group("/user")
	if deps.RequireAccessToken != nil {
		protectedUser.Use(deps.RequireAccessToken)
	}
	usershttp.RegisterProtectedRoutes(protectedUser, deps.UserHandler)
	authhttp.RegisterProtectedRoutes(protectedUser, deps.AuthHandler)

	publicPostcards := router.Group("/v1/postcards")
	if deps.OptionalAccessToken != nil {
		publicPostcards.Use(deps.OptionalAccessToken)
	}
	if deps.PostcardHandler != nil {
		postcardshttp.RegisterPublicRoutes(publicPostcards, deps.PostcardHandler)
	}
	if deps.MediaHandler != nil {
		mediahttp.RegisterPublicRoutes(publicPostcards, deps.MediaHandler)
	}

	protectedPostcards := router.Group("/v1/postcards")
	if deps.RequireAccessToken != nil {
		protectedPostcards.Use(deps.RequireAccessToken)
	}
	if deps.PostcardHandler != nil {
		postcardshttp.RegisterProtectedRoutes(protectedPostcards, deps.PostcardHandler)
	}
	if deps.MediaHandler != nil {
		mediahttp.RegisterProtectedRoutes(protectedPostcards, deps.MediaHandler)
	}

	return router
}

package app

import (
	authapp "chronote-refactor/internal/modules/auth/app"
	authhttp "chronote-refactor/internal/modules/auth/http"
	authinfra "chronote-refactor/internal/modules/auth/infra"
	healthapp "chronote-refactor/internal/modules/health/app"
	healthhttp "chronote-refactor/internal/modules/health/http"
	mediaapp "chronote-refactor/internal/modules/media/app"
	mediahttp "chronote-refactor/internal/modules/media/http"
	mediainfra "chronote-refactor/internal/modules/media/infra"
	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	postcardsapp "chronote-refactor/internal/modules/postcards/app"
	postcardshttp "chronote-refactor/internal/modules/postcards/http"
	postcardsinfra "chronote-refactor/internal/modules/postcards/infra"
	usersapp "chronote-refactor/internal/modules/users/app"
	usershttp "chronote-refactor/internal/modules/users/http"
	usersinfra "chronote-refactor/internal/modules/users/infra"
	platformauth "chronote-refactor/internal/platform/auth"
	platformconfig "chronote-refactor/internal/platform/config"
	platformdb "chronote-refactor/internal/platform/db"
	platformhttp "chronote-refactor/internal/platform/http"
	platformredis "chronote-refactor/internal/platform/redis"
	platforms3 "chronote-refactor/internal/platform/s3"

	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type App struct {
	router *gin.Engine
}

func New() (*App, error) {
	cfg, err := platformconfig.Load()
	if err != nil {
		return nil, err
	}

	database, err := platformdb.Open(cfg)
	if err != nil {
		return nil, err
	}

	redisClient := platformredis.NewClient(cfg)
	return newProductionApp(cfg, database, redisClient)
}

func NewTestApp() (*App, error) {
	passwordService := platformauth.PasswordService{}
	jwtService := platformauth.NewJWTService("test-signing-key", 7200, 1814400)
	userService := usersapp.NewService(nil, passwordService)
	authService := authapp.NewService(userService.Repository(), passwordService, tokenServiceAdapter{jwtService})
	mediaService := mediaapp.NewService(nil, nil, nil)
	postcardService := postcardsapp.NewService(nil, userService.Repository(), mediaService.Repository())
	analysisService := postcardaiapp.NewService(postcardaiapp.Dependencies{
		Enabled: false,
	})
	postcardService.SetAnalysisEnqueuer(analysisService)
	mediaService.SetAnalysisEnqueuer(analysisService)
	blacklist := authapp.NewMemoryBlacklist()
	authService.SetBlacklist(blacklist)
	middleware := authhttp.NewMiddleware(tokenServiceAdapter{jwtService}, blacklist)
	healthService := healthapp.NewService(nil, nil, nil)

	router := platformhttp.NewRouter(platformhttp.RouterDependencies{
		HealthHandler:       healthhttp.NewHandler(healthService),
		UserHandler:         usershttp.NewHandler(userService),
		AuthHandler:         authhttp.NewHandler(authService),
		PostcardHandler:     postcardshttp.NewHandler(postcardService, mediaService),
		MediaHandler:        mediahttp.NewHandler(mediaService, postcardService),
		RequireAccessToken:  middleware.RequireAccessToken(),
		OptionalAccessToken: middleware.OptionalAccessToken(),
	})

	return &App{router: router}, nil
}

func newProductionApp(cfg *platformconfig.Config, database *gorm.DB, redisClient *redislib.Client) (*App, error) {
	passwordService := platformauth.PasswordService{}
	jwtService := platformauth.NewJWTService(cfg.JWT.MySigningKey, cfg.JWT.AccessTokenExpire, cfg.JWT.RefreshTokenExpire)

	usersRepo := usersinfra.NewGormRepository(database)
	var blacklist authapp.TokenBlacklist = authapp.NewMemoryBlacklist()
	if redisClient != nil {
		blacklist = authinfra.NewRedisBlacklist(redisClient)
	}

	var s3Storage mediaapp.Storage
	var s3Checker healthapp.DependencyChecker
	if cfg.S3.Endpoint != "" && cfg.S3.BucketName != "" {
		s3Client, err := platforms3.NewClient(cfg)
		if err != nil {
			return nil, err
		}
		s3Storage = mediainfra.NewS3Storage(s3Client, cfg.S3.BucketName, cfg.S3.PublicBaseURL)
		s3Checker = platforms3.NewChecker(s3Client)
	}

	mediaRepo := mediainfra.NewGormRepository(database)
	mediaService := mediaapp.NewService(mediaRepo, s3Storage, mediainfra.NewImageProcessor())
	postcardService := postcardsapp.NewService(postcardsinfra.NewGormRepository(database), usersRepo, mediaRepo)
	analysisService, err := NewPostcardAIService(cfg, database, mediaRepo)
	if err != nil {
		return nil, err
	}
	postcardService.SetAnalysisEnqueuer(analysisService)
	mediaService.SetAnalysisEnqueuer(analysisService)
	userService := usersapp.NewService(usersRepo, passwordService)
	userService.SetAvatarStorage(s3Storage)
	authService := authapp.NewService(usersRepo, passwordService, tokenServiceAdapter{jwtService})
	authService.SetBlacklist(blacklist)
	middleware := authhttp.NewMiddleware(tokenServiceAdapter{jwtService}, blacklist)

	sqlDB, err := database.DB()
	if err != nil {
		return nil, err
	}
	healthService := healthapp.NewService(
		platformdb.NewChecker(sqlDB),
		platformredis.NewChecker(redisClient),
		s3Checker,
	)

	router := platformhttp.NewRouter(platformhttp.RouterDependencies{
		HealthHandler:       healthhttp.NewHandler(healthService),
		UserHandler:         usershttp.NewHandler(userService),
		AuthHandler:         authhttp.NewHandler(authService),
		PostcardHandler:     postcardshttp.NewHandler(postcardService, mediaService),
		MediaHandler:        mediahttp.NewHandler(mediaService, postcardService),
		RequireAccessToken:  middleware.RequireAccessToken(),
		OptionalAccessToken: middleware.OptionalAccessToken(),
	})

	return &App{router: router}, nil
}
func (a *App) Router() *gin.Engine {
	return a.router
}

type tokenServiceAdapter struct {
	*platformauth.JWTService
}

func (a tokenServiceAdapter) ParseToken(token string) (*authapp.TokenClaims, error) {
	claims, err := a.JWTService.ParseToken(token)
	if err != nil {
		return nil, err
	}
	return &authapp.TokenClaims{
		UserID:    claims.UserID,
		Name:      claims.Name,
		TokenType: claims.TokenType,
	}, nil
}

package integration

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	appplatform "chronote-refactor/internal/platform/app"
	platformauth "chronote-refactor/internal/platform/auth"
	platformconfig "chronote-refactor/internal/platform/config"
	platformdb "chronote-refactor/internal/platform/db"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type CutoverFixture struct {
	UserID            uint
	PublicPostcardID  uint
	PrivatePostcardID uint
	PublicMediaID     uint
}

type CutoverEnv struct {
	Config *platformconfig.Config
	DB     *gorm.DB
	Redis  *redislib.Client
}

func requireCutoverEnv(t *testing.T) *CutoverEnv {
	t.Helper()

	if os.Getenv("CHRONOTE_CUTOVER_TESTS") != "1" {
		t.Skip("set CHRONOTE_CUTOVER_TESTS=1 to run cutover integration tests")
	}

	cfg, err := platformconfig.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	db, err := platformdb.Open(cfg)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	redisClient := redislib.NewClient(&redislib.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	return &CutoverEnv{
		Config: cfg,
		DB:     db,
		Redis:  redisClient,
	}
}

func loadSupportedFixture(t *testing.T, env *CutoverEnv) CutoverFixture {
	t.Helper()

	resetSupportedData(t, env)

	passwords := platformauth.PasswordService{}
	hash, err := passwords.Hash("123456")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	now := time.Date(2026, 4, 22, 12, 0, 0, 0, time.FixedZone("CST", 8*3600))
	var userID uint
	if err := env.DB.Raw(`
		INSERT INTO users (username, display_name, email, password, avatar, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`, "fixture_user", "Fixture User", "fixture@example.com", hash, "https://cdn.example.com/avatar.png", now, now).Scan(&userID).Error; err != nil {
		t.Fatalf("insert user: %v", err)
	}

	var publicPostcardID uint
	if err := env.DB.Raw(`
		INSERT INTO postcards (title, content, visibility, author_id, created_at, updated_at)
		VALUES (?, ?::jsonb, ?, ?, ?, ?)
		RETURNING id
	`, "Fixture Public", `{"blocks":[{"type":"text","value":"public"}]}`, "public", userID, now, now).Scan(&publicPostcardID).Error; err != nil {
		t.Fatalf("insert public postcard: %v", err)
	}

	var privatePostcardID uint
	if err := env.DB.Raw(`
		INSERT INTO postcards (title, content, visibility, author_id, created_at, updated_at)
		VALUES (?, ?::jsonb, ?, ?, ?, ?)
		RETURNING id
	`, "Fixture Private", `{"blocks":[{"type":"text","value":"private"}]}`, "private", userID, now, now).Scan(&privatePostcardID).Error; err != nil {
		t.Fatalf("insert private postcard: %v", err)
	}

	var publicMediaID uint
	if err := env.DB.Raw(`
		INSERT INTO postcard_media (
			postcard_id, media_type, url, thumbnail_url, oss_key, thumbnail_oss_key,
			original_width, original_height, duration, file_size, position, media_group, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`, publicPostcardID, "image", "https://cdn.example.com/postcards/public.jpg", "https://cdn.example.com/postcards/public-thumb.jpg", "postcards/public.jpg", "postcards/public-thumb.jpg", 1280, 720, 0, 2048, 1, "gallery", now, now).Scan(&publicMediaID).Error; err != nil {
		t.Fatalf("insert public media: %v", err)
	}

	return CutoverFixture{
		UserID:            userID,
		PublicPostcardID:  publicPostcardID,
		PrivatePostcardID: privatePostcardID,
		PublicMediaID:     publicMediaID,
	}
}

func newProductionApp(t *testing.T) *appplatform.App {
	t.Helper()

	app, err := appplatform.New()
	if err != nil {
		t.Fatalf("new production app: %v", err)
	}
	return app
}

func loginFixtureUser(t *testing.T, app *appplatform.App) string {
	t.Helper()

	body := bytes.NewBufferString(`{"email":"fixture@example.com","password":"123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/user/login", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected fixture login 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	return extractAccessToken(t, rec.Body.Bytes())
}

func resetSupportedData(t *testing.T, env *CutoverEnv) {
	t.Helper()

	sqlDB, err := env.DB.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
		_ = env.Redis.Close()
	})

	statements := []string{
		`DROP TABLE IF EXISTS postcard_media`,
		`DROP TABLE IF EXISTS postcards`,
		`DROP TABLE IF EXISTS users`,
		`CREATE TABLE users (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(50) UNIQUE NOT NULL,
			display_name VARCHAR(100),
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			avatar VARCHAR(500),
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE postcards (
			id BIGSERIAL PRIMARY KEY,
			title VARCHAR(200) NOT NULL,
			content JSONB NOT NULL,
			visibility VARCHAR(20) NOT NULL,
			author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE postcard_media (
			id BIGSERIAL PRIMARY KEY,
			postcard_id BIGINT NOT NULL REFERENCES postcards(id) ON DELETE CASCADE,
			media_type VARCHAR(20) NOT NULL,
			url VARCHAR(500) NOT NULL,
			thumbnail_url VARCHAR(500),
			oss_key VARCHAR(500) NOT NULL,
			thumbnail_oss_key VARCHAR(500),
			original_width INTEGER,
			original_height INTEGER,
			duration INTEGER,
			file_size BIGINT NOT NULL,
			position INTEGER NOT NULL,
			media_group VARCHAR(50) NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)`,
	}

	for _, statement := range statements {
		if err := env.DB.Exec(statement).Error; err != nil {
			t.Fatalf("exec schema statement: %v", err)
		}
	}

	if err := env.Redis.FlushDB(context.Background()).Err(); err != nil {
		t.Fatalf("flush redis: %v", err)
	}
}

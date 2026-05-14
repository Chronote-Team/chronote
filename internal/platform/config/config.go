package config

import (
	"errors"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Database struct {
		Host     string
		Port     string
		User     string
		Password string
		DBName   string
		SSLmode  string
		TimeZone string
	}
	JWT struct {
		MySigningKey       string
		AccessTokenExpire  int64
		RefreshTokenExpire int64
	}
	Redis struct {
		Host     string
		Port     string
		Password string
		DB       int
	}
	S3 struct {
		Endpoint        string
		AccessKeyID     string
		SecretAccessKey string
		BucketName      string
		Region          string
		UseSSL          bool
	}
	Media struct {
		MaxImageSize int64
		MaxVideoSize int64
		MaxAudioSize int64
	}
	AI struct {
		Enabled      bool
		Provider     string
		Endpoint     string
		EndpointType string
		Model        string
		Timeout      int64
		OpenAIAPIKey string
	}
}

var AppConfig *Config

func Load() (*Config, error) {
	viper.Reset()
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config"
	}
	viper.AddConfigPath(configPath)

	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.timezone", "Asia/Shanghai")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("s3.region", "us-east-1")
	viper.SetDefault("s3.usessl", false)
	viper.SetDefault("jwt.accesstokenexpire", 7200)
	viper.SetDefault("jwt.refreshtokenexpire", 1814400)
	viper.SetDefault("media.maximagesize", int64(10*1024*1024))
	viper.SetDefault("media.maxvideosize", int64(200*1024*1024))
	viper.SetDefault("media.maxaudiosize", int64(50*1024*1024))
	viper.SetDefault("ai.enabled", false)
	viper.SetDefault("ai.provider", "openai")
	viper.SetDefault("ai.endpointtype", "responses")
	viper.SetDefault("ai.timeout", int64(30))

	bindEnv("database.host", "POSTGRES_HOST")
	bindEnv("database.port", "POSTGRES_PORT")
	bindEnv("database.user", "POSTGRES_USER")
	bindEnv("database.password", "POSTGRES_PASSWORD")
	bindEnv("database.dbname", "POSTGRES_DB")
	bindEnv("database.sslmode", "POSTGRES_SSLMODE")
	bindEnv("database.timezone", "POSTGRES_TIMEZONE")

	bindEnv("jwt.mysigningkey", "JWT_SIGNING_KEY")
	bindEnv("jwt.accesstokenexpire", "ACCESS_TOKEN_EXPIRE")
	bindEnv("jwt.refreshtokenexpire", "REFRESH_TOKEN_EXPIRE")

	bindEnv("redis.host", "REDIS_HOST")
	bindEnv("redis.port", "REDIS_PORT")
	bindEnv("redis.password", "REDIS_PASSWORD")
	bindEnv("redis.db", "REDIS_DB")

	bindEnv("s3.endpoint", "S3_ENDPOINT")
	bindEnv("s3.accesskeyid", "S3_ACCESS_KEY")
	bindEnv("s3.secretaccesskey", "S3_SECRET_KEY")
	bindEnv("s3.bucketname", "S3_BUCKET")
	bindEnv("s3.region", "S3_REGION")
	bindEnv("s3.usessl", "S3_USE_SSL")

	bindEnv("media.maximagesize", "MEDIA_MAX_IMAGE_SIZE")
	bindEnv("media.maxvideosize", "MEDIA_MAX_VIDEO_SIZE")
	bindEnv("media.maxaudiosize", "MEDIA_MAX_AUDIO_SIZE")

	bindEnv("ai.enabled", "AI_ENABLED")
	bindEnv("ai.provider", "AI_PROVIDER")
	bindEnv("ai.endpoint", "AI_ENDPOINT")
	bindEnv("ai.endpointtype", "AI_ENDPOINT_TYPE")
	bindEnv("ai.model", "AI_MODEL")
	bindEnv("ai.timeout", "AI_TIMEOUT")
	bindEnv("ai.openaiapikey", "OPENAI_API_KEY")

	var configFileNotFound viper.ConfigFileNotFoundError
	if err := viper.ReadInConfig(); err != nil && !errors.As(err, &configFileNotFound) {
		return nil, err
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}
	AppConfig = cfg
	return cfg, nil
}

func bindEnv(key, env string) {
	_ = viper.BindEnv(key, env)
}

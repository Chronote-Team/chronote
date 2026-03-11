package config

import (
	"errors"
	"log"
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
}

var AppConfig *Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	configPath := os.Getenv("CONFIG_PATH") // Deploy Environment in CONFIG_PATH, need to be set
	if configPath == "" {
		configPath = "./config" // Develop Environment in "./config"
	}
	viper.AddConfigPath(configPath)

	// Defaults keep local dev ergonomic when some env vars are omitted.
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.timezone", "Asia/Shanghai")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("s3.region", "us-east-1")
	viper.SetDefault("s3.usessl", false)
	viper.SetDefault("jwt.accesstokenexpire", 7200)
	viper.SetDefault("jwt.refreshtokenexpire", 1814400)

	// Database
	mustBindEnv("database.host", "POSTGRES_HOST")
	mustBindEnv("database.port", "POSTGRES_PORT")
	mustBindEnv("database.user", "POSTGRES_USER")
	mustBindEnv("database.password", "POSTGRES_PASSWORD")
	mustBindEnv("database.dbname", "POSTGRES_DB")
	mustBindEnv("database.sslmode", "POSTGRES_SSLMODE")
	mustBindEnv("database.timezone", "POSTGRES_TIMEZONE")

	// JWT
	mustBindEnv("jwt.mysigningkey", "JWT_SIGNING_KEY")
	mustBindEnv("jwt.accesstokenexpire", "ACCESS_TOKEN_EXPIRE")
	mustBindEnv("jwt.refreshtokenexpire", "REFRESH_TOKEN_EXPIRE")

	// Redis
	mustBindEnv("redis.host", "REDIS_HOST")
	mustBindEnv("redis.port", "REDIS_PORT")
	mustBindEnv("redis.password", "REDIS_PASSWORD")
	mustBindEnv("redis.db", "REDIS_DB")

	// S3
	mustBindEnv("s3.endpoint", "S3_ENDPOINT")
	mustBindEnv("s3.accesskeyid", "S3_ACCESS_KEY")
	mustBindEnv("s3.secretaccesskey", "S3_SECRET_KEY")
	mustBindEnv("s3.bucketname", "S3_BUCKET")
	mustBindEnv("s3.region", "S3_REGION")
	mustBindEnv("s3.usessl", "S3_USE_SSL")

	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFound) {
			log.Fatalf("Error in Reading config file, %v", err)
		}
		log.Printf("Config file not found in %s, reading settings from environment variables", configPath)
	}

	AppConfig = &Config{}

	if err := viper.Unmarshal(AppConfig); err != nil {
		log.Fatalf("Failed to decode config file, %v", err)
	}

	validateRequiredConfig()
}

func mustBindEnv(key, env string) {
	if err := viper.BindEnv(key, env); err != nil {
		log.Fatalf("Failed to bind env %s to %s, %v", env, key, err)
	}
}

func validateRequiredConfig() {
	required := map[string]string{
		"database.host":      AppConfig.Database.Host,
		"database.port":      AppConfig.Database.Port,
		"database.user":      AppConfig.Database.User,
		"database.password":  AppConfig.Database.Password,
		"database.dbname":    AppConfig.Database.DBName,
		"jwt.mysigningkey":   AppConfig.JWT.MySigningKey,
		"redis.host":         AppConfig.Redis.Host,
		"redis.port":         AppConfig.Redis.Port,
		"s3.endpoint":        AppConfig.S3.Endpoint,
		"s3.accesskeyid":     AppConfig.S3.AccessKeyID,
		"s3.secretaccesskey": AppConfig.S3.SecretAccessKey,
		"s3.bucketname":      AppConfig.S3.BucketName,
	}

	for key, value := range required {
		if strings.TrimSpace(value) == "" {
			log.Fatalf("Missing required config: %s (set env var or provide config.yml)", key)
		}
	}
}

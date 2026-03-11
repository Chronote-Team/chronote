package config

import (
	"log"
	"os"

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

	configPath := os.Getenv("CONFIG_PATH") // Deploy Environment in CONFIG_PATH, need to be set
	if configPath == "" {
		configPath = "./config" // Develop Environment in "./config"
	}
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error in Reading config file, %v", err)
	}

	AppConfig = &Config{}

	if err := viper.Unmarshal(AppConfig); err != nil {
		log.Fatalf("Failed to decode config file, %v", err)
	}
}

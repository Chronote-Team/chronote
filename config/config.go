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
		dbName   string
		sslmode  string
		TimeZone string
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

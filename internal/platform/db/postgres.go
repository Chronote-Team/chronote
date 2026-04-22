package db

import (
	"fmt"

	"chronote-refactor/internal/platform/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLmode,
		cfg.Database.TimeZone,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

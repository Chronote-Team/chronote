package config

import (
	"chronote/global"
	"chronote/models"

	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func buildDSN(host, user, password, dbname, port, sslmode, TimeZone string) string {
	template := "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s"
	return fmt.Sprintf(template, host, user, password, dbname, port, sslmode, TimeZone)
}

func InitDB() {
	dsn := buildDSN(AppConfig.Database.Host, AppConfig.Database.User, AppConfig.Database.Password,
		AppConfig.Database.DBName, AppConfig.Database.Port, AppConfig.Database.SSLmode,
		AppConfig.Database.TimeZone)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect to database,%v", err)
	}
	err = db.AutoMigrate(
		&models.User{},
	)
	if err != nil {
		log.Fatalf("Failed to auto migrate database, %v", err)
	}
	log.Println("Database connected and migrated successfully.")

	global.Db = db
}

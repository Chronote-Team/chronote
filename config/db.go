package config

import (
	"chronote/global"

	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func buildDSN(user, password, host, port, dbname, sslmode, TimeZone string) string {
	template := "host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s"
	return fmt.Sprintf(template, host, user, password, dbname, port, sslmode, TimeZone)
}

func initDB() {
	dsn := buildDSN(AppConfig.Database.Host, AppConfig.Database.User, AppConfig.Database.Password, AppConfig.Database.dbName, AppConfig.Database.Port, AppConfig.Database.sslmode, AppConfig.Database.TimeZone)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Failed to connect to database,%v", err)
	}

	global.Db = db
}

package models

type User struct {
	BaseModel
	Username    string `gorm:"type:varchar(50);uniqueIndex;not null"`
	DisplayName string `gorm:"type:varchar(100)"`
	Email       string `gorm:"type:varchar(255);uniqueIndex;not null"`
	Password    string `gorm:"type:varchar(255);not null"`
	Avatar      string `gorm:"type:varchar(500)"`
}

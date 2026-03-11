package models

type User struct {
	BaseModel
	Username    string `gorm:"type:varchar(100);unique;not null" json:"username" binding:"required"`
	DisplayName string `gorm:"type:varchar(100)" json:"display_name"`
	Email       string `gorm:"type:varchar(255);unique;not null" json:"email" binding:"required,email"`
	Password    string `gorm:"type:varchar(255);not null" json:"password" binding:"required,min=6"`
	Avatar      string `gorm:"type:varchar(500)" json:"avatar,omitempty"`
}

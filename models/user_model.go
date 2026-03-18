package models

type User struct {
	BaseModel
	Username    string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	DisplayName string `gorm:"type:varchar(100)" json:"display_name"`
	Email       string `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password    string `gorm:"type:varchar(255);not null" json:"password" binding:"required,min=6"`
	Avatar      string `gorm:"type:varchar(500)" json:"avatar,omitempty"`
}

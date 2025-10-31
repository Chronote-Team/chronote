package models

import (
	"gorm.io/gorm"
)

type User struct {    
	gorm.Model    
	Username string `gorm:"type:varchar(100);unique;not null" json:"username" binding:"required"`    
	Email    string `gorm:"type:varchar(255);unique;not null" json:"email" binding:"required,email"`    
	Password string `gorm:"type:varchar(255);not null" json:"password" binding:"required,min=6"`
}

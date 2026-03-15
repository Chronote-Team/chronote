package models

import "gorm.io/datatypes"

type Postcard struct {
	BaseModel
	Title      string          `gorm:"type:varchar(200);not null" json:"title"`
	Content    datatypes.JSON  `gorm:"type:jsonb" json:"content"`
	Visibility string          `gorm:"type:varchar(20);default:private" json:"visibility"`
	AuthorID   uint            `gorm:"not null;index" json:"author_id"`
	Author     *User           `gorm:"foreignKey:AuthorID" json:"-"`
	Medias     []PostcardMedia `gorm:"foreignKey:PostcardID" json:"medias,omitempty"`
}

type PostcardMedia struct {
	BaseModel
	PostcardID      uint   `gorm:"not null;index" json:"postcard_id"`
	MediaType       string `gorm:"type:varchar(20);not null" json:"type"`
	OSSKey          string `gorm:"type:varchar(500);not null" json:"oss_key"`
	URL             string `gorm:"type:varchar(500)" json:"url"`
	ThumbnailURL    string `gorm:"type:varchar(500)" json:"thumbnail_url,omitempty"`
	ThumbnailOSSKey string `gorm:"type:varchar(500)" json:"-"`
	OriginalWidth   int    `json:"original_width,omitempty"`
	OriginalHeight  int    `json:"original_height,omitempty"`
	Duration        int    `json:"duration,omitempty"`
	FileSize        int64  `json:"file_size"`
	Position        int    `gorm:"default:0" json:"position"`
	MediaGroup      string `gorm:"type:varchar(50);default:gallery" json:"group"`
}

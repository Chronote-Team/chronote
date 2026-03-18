package models

import "gorm.io/datatypes"

type Postcard struct {
	BaseModel
	Title      string          `gorm:"type:varchar(200);not null;check:postcards_title_not_blank,title <> ''"`
	Content    datatypes.JSON  `gorm:"type:jsonb;not null"`
	Visibility string          `gorm:"type:varchar(20);not null;default:private;check:postcards_visibility_valid,visibility IN ('public','private')"`
	AuthorID   uint            `gorm:"not null;index"`
	Author     *User           `gorm:"foreignKey:AuthorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Medias     []PostcardMedia `gorm:"foreignKey:PostcardID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type PostcardMedia struct {
	BaseModel
	PostcardID      uint   `gorm:"not null;index;uniqueIndex:idx_postcard_media_position"`
	MediaType       string `gorm:"type:varchar(20);not null;check:postcard_media_type_valid,media_type IN ('image','video','audio')"`
	OSSKey          string `gorm:"type:varchar(500);not null"`
	URL             string `gorm:"type:varchar(500);not null"`
	ThumbnailURL    string `gorm:"type:varchar(500)"`
	ThumbnailOSSKey string `gorm:"type:varchar(500)"`
	OriginalWidth   int
	OriginalHeight  int
	Duration        int
	FileSize        int64  `gorm:"not null;check:postcard_media_file_size_positive,file_size > 0"`
	Position        int    `gorm:"not null;default:1;uniqueIndex:idx_postcard_media_position;check:postcard_media_position_positive,position > 0"`
	MediaGroup      string `gorm:"type:varchar(50);not null;default:gallery;check:postcard_media_group_valid,media_group IN ('header','gallery','bgm')"`
}

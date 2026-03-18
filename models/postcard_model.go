package models

import "gorm.io/datatypes"

type Postcard struct {
	BaseModel
	Title      string          `gorm:"type:varchar(200);not null;check:postcards_title_not_blank,title <> ''" json:"title"`
	Content    datatypes.JSON  `gorm:"type:jsonb;not null" json:"content"`
	Visibility string          `gorm:"type:varchar(20);not null;default:private;check:postcards_visibility_valid,visibility IN ('public','private')" json:"visibility"`
	AuthorID   uint            `gorm:"not null;index" json:"author_id"`
	Author     *User           `gorm:"foreignKey:AuthorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"-"`
	Medias     []PostcardMedia `gorm:"foreignKey:PostcardID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"medias,omitempty"`
}

type PostcardMedia struct {
	BaseModel
	PostcardID      uint   `gorm:"not null;index;uniqueIndex:idx_postcard_media_position" json:"postcard_id"`
	MediaType       string `gorm:"type:varchar(20);not null;check:postcard_media_type_valid,media_type IN ('image','video','audio')" json:"type"`
	OSSKey          string `gorm:"type:varchar(500);not null" json:"oss_key"`
	URL             string `gorm:"type:varchar(500);not null" json:"url"`
	ThumbnailURL    string `gorm:"type:varchar(500)" json:"thumbnail_url,omitempty"`
	ThumbnailOSSKey string `gorm:"type:varchar(500)" json:"-"`
	OriginalWidth   int    `json:"original_width,omitempty"`
	OriginalHeight  int    `json:"original_height,omitempty"`
	Duration        int    `json:"duration,omitempty"`
	FileSize        int64  `gorm:"not null;check:postcard_media_file_size_positive,file_size > 0" json:"file_size"`
	Position        int    `gorm:"not null;default:1;uniqueIndex:idx_postcard_media_position;check:postcard_media_position_positive,position > 0" json:"position"`
	MediaGroup      string `gorm:"type:varchar(50);not null;default:gallery;check:postcard_media_group_valid,media_group IN ('header','gallery','bgm')" json:"group"`
}

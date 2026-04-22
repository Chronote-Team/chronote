package domain

import "time"

type Media struct {
	ID                  uint
	PostcardID          uint
	Type                string
	URL                 string
	ThumbnailURL        string
	StorageKey          string
	ThumbnailStorageKey string
	OriginalWidth       int
	OriginalHeight      int
	Duration            int
	FileSize            int64
	Position            int
	Group               string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

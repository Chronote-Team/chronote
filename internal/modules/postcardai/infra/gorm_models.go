package infra

import (
	"time"
)

type AnalysisJobModel struct {
	ID              uint `gorm:"primaryKey"`
	PostcardID      uint
	PostcardVersion string
	Status          string
	Attempts        int
	NextRunAt       *time.Time
	LockedAt        *time.Time
	LastErrorCode   string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (AnalysisJobModel) TableName() string { return "ai_analysis_jobs" }

type MediaAnalysisModel struct {
	ID            uint `gorm:"primaryKey"`
	MediaID       uint
	MediaVersion  string
	PromptVersion string
	SchemaVersion string
	ModelVersion  string
	Status        string
	ResultJSON    []byte
	Confidence    float64
	Uncertainty   string
	ErrorCode     string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (MediaAnalysisModel) TableName() string { return "media_ai_analysis" }

type PostcardAnalysisModel struct {
	ID              uint `gorm:"primaryKey"`
	PostcardID      uint
	PostcardVersion string
	PromptVersion   string
	SchemaVersion   string
	ModelVersion    string
	Status          string
	ResultJSON      []byte
	Confidence      float64
	Uncertainty     string
	ErrorCode       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (PostcardAnalysisModel) TableName() string { return "postcard_ai_analysis" }

package domain

import (
	"encoding/json"
	"time"
)

type MediaAnalysis struct {
	ID            uint
	MediaID       uint
	MediaVersion  string
	PromptVersion string
	SchemaVersion string
	ModelVersion  string
	Status        AnalysisStatus
	Result        json.RawMessage
	Confidence    float64
	Uncertainty   string
	ErrorCode     ProviderErrorCode
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (a MediaAnalysis) VersionKey() VersionKey {
	return VersionKey{
		ResourceID:      a.MediaID,
		ResourceVersion: a.MediaVersion,
		PromptVersion:   a.PromptVersion,
		SchemaVersion:   a.SchemaVersion,
		ModelVersion:    a.ModelVersion,
	}
}

type PostcardAnalysis struct {
	ID              uint
	PostcardID      uint
	PostcardVersion string
	PromptVersion   string
	SchemaVersion   string
	ModelVersion    string
	Status          AnalysisStatus
	Result          json.RawMessage
	Confidence      float64
	Uncertainty     string
	ErrorCode       ProviderErrorCode
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (a PostcardAnalysis) VersionKey() VersionKey {
	return VersionKey{
		ResourceID:      a.PostcardID,
		ResourceVersion: a.PostcardVersion,
		PromptVersion:   a.PromptVersion,
		SchemaVersion:   a.SchemaVersion,
		ModelVersion:    a.ModelVersion,
	}
}

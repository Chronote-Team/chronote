package domain

import "time"

type AnalysisJob struct {
	ID              uint
	PostcardID      uint
	PostcardVersion string
	Status          AnalysisStatus
	Attempts        int
	NextRunAt       *time.Time
	LockedAt        *time.Time
	LastErrorCode   ProviderErrorCode
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewAnalysisJob(postcardID uint, postcardVersion string, now time.Time) AnalysisJob {
	return AnalysisJob{
		PostcardID:      postcardID,
		PostcardVersion: postcardVersion,
		Status:          StatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func (j AnalysisJob) IsActive() bool {
	return j.Status == StatusPending || j.Status == StatusProcessing
}

func (j AnalysisJob) IsCurrentFor(postcardVersion string) bool {
	return j.PostcardVersion == postcardVersion
}

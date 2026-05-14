package domain

type AnalysisStatus string

const (
	StatusPending     AnalysisStatus = "pending"
	StatusProcessing  AnalysisStatus = "processing"
	StatusSucceeded   AnalysisStatus = "succeeded"
	StatusFailed      AnalysisStatus = "failed"
	StatusUnavailable AnalysisStatus = "unavailable"
	StatusStale       AnalysisStatus = "stale"
)

func CanTransition(from, to AnalysisStatus) bool {
	switch from {
	case StatusPending:
		return to == StatusProcessing || to == StatusStale
	case StatusProcessing:
		return to == StatusSucceeded || to == StatusFailed || to == StatusUnavailable || to == StatusStale || to == StatusPending
	case StatusFailed, StatusUnavailable:
		return to == StatusPending
	default:
		return false
	}
}

func (s AnalysisStatus) Successful() bool {
	return s == StatusSucceeded
}

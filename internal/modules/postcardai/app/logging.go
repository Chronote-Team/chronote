package app

import (
	"fmt"
	"strings"

	"chronote-refactor/internal/modules/postcardai/domain"
)

type ProviderCallLog struct {
	JobID           uint
	PostcardID      uint
	MediaIDs        []uint
	ModelVersion    string
	PromptVersion   string
	SchemaVersion   string
	Status          domain.AnalysisStatus
	ErrorCode       domain.ProviderErrorCode
	LatencyMS       int64
	RetryCount      int
	RawPostcardText string
	SignedURL       string
	APIKey          string
	ProviderPayload string
}

func SafeLogFields(input ProviderCallLog) []string {
	fields := []string{
		fmt.Sprintf("job_id=%d", input.JobID),
		fmt.Sprintf("postcard_id=%d", input.PostcardID),
		"media_ids=" + formatIDs(input.MediaIDs),
		"model=" + input.ModelVersion,
		"prompt_version=" + input.PromptVersion,
		"schema_version=" + input.SchemaVersion,
		"status=" + string(input.Status),
		"error_code=" + string(input.ErrorCode),
		fmt.Sprintf("latency_ms=%d", input.LatencyMS),
		fmt.Sprintf("retry_count=%d", input.RetryCount),
	}
	return fields
}

func formatIDs(ids []uint) string {
	if len(ids) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, fmt.Sprintf("%d", id))
	}
	return "[" + strings.Join(parts, ",") + "]"
}

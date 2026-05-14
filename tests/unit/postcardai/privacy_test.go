package postcardai_test

import (
	"strings"
	"testing"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
)

func TestSafeLogFieldsExcludeSensitiveValues(t *testing.T) {
	fields := postcardaiapp.SafeLogFields(postcardaiapp.ProviderCallLog{
		JobID:           1,
		PostcardID:      2,
		MediaIDs:        []uint{3},
		ModelVersion:    "test-model",
		PromptVersion:   "prompt-v1",
		SchemaVersion:   "schema-v1",
		Status:          domain.StatusFailed,
		ErrorCode:       domain.ErrorProviderUnavailable,
		RawPostcardText: "secret diary",
		SignedURL:       "https://signed.example.com/private?token=secret",
		APIKey:          "sk-secret",
		ProviderPayload: `{"input":"secret diary"}`,
	})
	rendered := strings.Join(fields, " ")
	for _, forbidden := range []string{"secret diary", "signed.example.com", "sk-secret", "ProviderPayload"} {
		if strings.Contains(rendered, forbidden) {
			t.Fatalf("safe log fields leaked %q in %q", forbidden, rendered)
		}
	}
}

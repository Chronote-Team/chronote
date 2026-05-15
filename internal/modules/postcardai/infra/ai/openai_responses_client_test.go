package ai

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
)

func TestOpenAIResponsesClientParsesOutputContentText(t *testing.T) {
	result, err := parseResponsesResult([]byte(`{
		"output": [{
			"content": [{
				"type": "output_text",
				"text": "{\"summary\":\"lunch\",\"suggested_title\":\"Lunch\",\"confidence\":0.91,\"uncertainty\":\"low\"}"
			}]
		}]
	}`))
	if err != nil {
		t.Fatalf("parseResponsesResult returned error: %v", err)
	}
	if result.Confidence != 0.91 {
		t.Fatalf("Confidence = %v, want 0.91", result.Confidence)
	}
	if result.Uncertainty != "low" {
		t.Fatalf("Uncertainty = %q, want low", result.Uncertainty)
	}
	if !strings.Contains(string(result.JSON), `"suggested_title":"Lunch"`) {
		t.Fatalf("unexpected JSON: %s", string(result.JSON))
	}
}

func TestOpenAIResponsesClientParsesOutputTextAndFencedJSON(t *testing.T) {
	raw := "{\n" +
		"\"output_text\": \"```json\\n{\\\"summary\\\":\\\"trip\\\",\\\"suggested_title\\\":\\\"Trip\\\",\\\"confidence\\\":0.8}\\n```\"\n" +
		"}"
	result, err := parseResponsesResult([]byte(raw))
	if err != nil {
		t.Fatalf("parseResponsesResult returned error: %v", err)
	}
	if !strings.Contains(string(result.JSON), `"summary":"trip"`) {
		t.Fatalf("unexpected JSON: %s", string(result.JSON))
	}
}

func TestOpenAIResponsesClientRejectsMalformedOutput(t *testing.T) {
	_, err := parseResponsesResult([]byte(`{"output_text":"not json"}`))
	if err == nil {
		t.Fatal("expected malformed output error")
	}
}

func TestOpenAIResponsesClientClassifiesUnavailable(t *testing.T) {
	client := &OpenAIResponsesClient{
		apiKey:   "test-key",
		model:    "test-model",
		endpoint: "https://example.invalid/responses",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Status:     "429 Too Many Requests",
				Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				Header:     make(http.Header),
			}, nil
		})},
	}
	_, err := client.AnalyzePostcard(context.Background(), postcardaiapp.PostcardUnderstandingRequest{})
	var providerErr domain.ProviderError
	if !errors.As(err, &providerErr) {
		t.Fatalf("expected ProviderError, got %T %v", err, err)
	}
	if providerErr.Code != domain.ErrorProviderUnavailable {
		t.Fatalf("Code = %q, want %q", providerErr.Code, domain.ErrorProviderUnavailable)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

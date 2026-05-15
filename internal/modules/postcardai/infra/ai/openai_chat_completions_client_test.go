package ai

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
)

func TestOpenAIChatCompletionsClientParsesMessageContent(t *testing.T) {
	result, err := parseChatCompletionsResult([]byte(`{
		"choices": [{
			"message": {
				"role": "assistant",
				"content": "{\"summary\":\"lunch\",\"suggested_title\":\"Lunch\",\"confidence\":0.91,\"uncertainty\":\"low\"}"
			}
		}]
	}`))
	if err != nil {
		t.Fatalf("parseChatCompletionsResult returned error: %v", err)
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

func TestOpenAIChatCompletionsClientParsesFencedJSON(t *testing.T) {
	raw := "{\n" +
		"\"choices\": [{\"message\": {" +
		"\"content\": \"```json\\n{\\\"summary\\\":\\\"trip\\\",\\\"suggested_title\\\":\\\"Trip\\\",\\\"confidence\\\":0.8}\\n```\"" +
		"}}]\n" +
		"}"
	result, err := parseChatCompletionsResult([]byte(raw))
	if err != nil {
		t.Fatalf("parseChatCompletionsResult returned error: %v", err)
	}
	if !strings.Contains(string(result.JSON), `"summary":"trip"`) {
		t.Fatalf("unexpected JSON: %s", string(result.JSON))
	}
}

func TestOpenAIChatCompletionsClientUsesChatEndpointPayload(t *testing.T) {
	var gotPath string
	var gotPayload string
	client := &OpenAIChatCompletionsClient{
		apiKey:   "test-key",
		model:    "test-model",
		endpoint: "https://example.test/v1/chat/completions",
		httpClient: &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			gotPath = r.URL.Path
			raw, _ := io.ReadAll(r.Body)
			gotPayload = string(raw)
			return jsonResponse(http.StatusOK, `{"choices":[{"message":{"content":"{\"summary\":\"ok\",\"suggested_title\":\"OK\",\"confidence\":0.7}"}}]}`)
		})},
	}
	_, err := client.AnalyzePostcard(context.Background(), postcardaiapp.PostcardUnderstandingRequest{
		Postcard: postcardaiapp.PostcardSnapshot{Content: []byte(`{"text":"hello"}`)},
	})
	if err != nil {
		t.Fatalf("AnalyzePostcard returned error: %v", err)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("path = %q, want /v1/chat/completions", gotPath)
	}
	for _, expected := range []string{`"messages"`, `Postcard content`, `{\"text\":\"hello\"}`} {
		if !strings.Contains(gotPayload, expected) {
			t.Fatalf("payload missing %q: %s", expected, gotPayload)
		}
	}
}

func TestOpenAIChatCompletionsClientRejectsMalformedOutput(t *testing.T) {
	_, err := parseChatCompletionsResult([]byte(`{"choices":[{"message":{"content":"not json"}}]}`))
	if err == nil {
		t.Fatal("expected malformed output error")
	}
}

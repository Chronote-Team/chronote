package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
)

const maxResponseBodyBytes = 2 * 1024 * 1024

type OpenAIResponsesClient struct {
	apiKey     string
	model      string
	endpoint   string
	httpClient *http.Client
}

func NewOpenAIResponsesClient(apiKey, model, endpoint string, timeout time.Duration) *OpenAIResponsesClient {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/responses"
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &OpenAIResponsesClient{
		apiKey:     apiKey,
		model:      model,
		endpoint:   endpoint,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *OpenAIResponsesClient) AnalyzeImage(ctx context.Context, req postcardaiapp.ImageUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	if c.apiKey == "" {
		return nil, domain.ProviderError{Code: domain.ErrorProviderUnavailable, Err: errors.New("missing API key")}
	}
	payload := map[string]any{
		"model": c.model,
		"input": []map[string]any{{
			"role": "user",
			"content": []map[string]string{
				{"type": "input_text", "text": ImagePromptV1()},
				{"type": "input_image", "image_url": req.SignedURL},
			},
		}},
	}
	return c.do(ctx, payload)
}

func (c *OpenAIResponsesClient) RepairImage(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return c.repair(ctx, req)
}

func (c *OpenAIResponsesClient) AnalyzePostcard(ctx context.Context, req postcardaiapp.PostcardUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	if c.apiKey == "" {
		return nil, domain.ProviderError{Code: domain.ErrorProviderUnavailable, Err: errors.New("missing API key")}
	}
	payload := map[string]any{
		"model": c.model,
		"input": []map[string]any{{
			"role": "user",
			"content": []map[string]string{
				{"type": "input_text", "text": PostcardPromptV1()},
				{"type": "input_text", "text": string(req.Postcard.Content)},
			},
		}},
	}
	return c.do(ctx, payload)
}

func (c *OpenAIResponsesClient) RepairPostcard(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return c.repair(ctx, req)
}

func (c *OpenAIResponsesClient) repair(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	if c.apiKey == "" {
		return nil, domain.ProviderError{Code: domain.ErrorProviderUnavailable, Err: errors.New("missing API key")}
	}
	payload := map[string]any{
		"model": c.model,
		"input": "Repair this JSON for schema " + req.SchemaVersion + ": " + string(req.Original),
	}
	return c.do(ctx, payload)
}

func (c *OpenAIResponsesClient) do(ctx context.Context, payload any) (*postcardaiapp.AIResult, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	request.Header.Set("Content-Type", "application/json")
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, domain.ProviderError{Code: domain.ErrorProviderUnavailable, Err: err}
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500 {
		return nil, domain.ProviderError{Code: domain.ErrorProviderUnavailable, Err: errors.New(response.Status)}
	}
	if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusBadRequest {
		return nil, domain.ProviderError{Code: domain.ErrorProviderRefused, Err: errors.New(response.Status)}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, domain.ProviderError{Code: domain.ErrorProviderRefused, Err: errors.New(response.Status)}
	}
	raw, err := readBounded(response.Body, maxResponseBodyBytes)
	if err != nil {
		return nil, domain.ProviderError{Code: domain.ErrorMalformedOutput, Err: err}
	}
	result, err := parseResponsesResult(raw)
	if err != nil {
		return nil, domain.ProviderError{Code: domain.ErrorMalformedOutput, Err: err}
	}
	return result, nil
}

func readBounded(reader io.Reader, limit int64) ([]byte, error) {
	limited := io.LimitReader(reader, limit+1)
	raw, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > limit {
		return nil, fmt.Errorf("response body exceeds %d bytes", limit)
	}
	return raw, nil
}

func parseResponsesResult(raw []byte) (*postcardaiapp.AIResult, error) {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	text, ok := extractResponseText(payload)
	if !ok {
		return nil, errors.New("response did not include supported text output")
	}
	text = trimJSONFence(text)
	if !json.Valid([]byte(text)) {
		return nil, errors.New("response output is not valid JSON")
	}
	result := json.RawMessage(text)
	confidence, uncertainty := extractResultMetadata(result)
	return &postcardaiapp.AIResult{
		JSON:        cloneRaw(result),
		Confidence:  confidence,
		Uncertainty: uncertainty,
	}, nil
}

func extractResponseText(payload map[string]any) (string, bool) {
	if text, ok := stringField(payload, "output_text"); ok {
		return text, true
	}
	output, ok := payload["output"].([]any)
	if !ok {
		return "", false
	}
	for _, item := range output {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if text, ok := stringField(itemMap, "output_text"); ok {
			return text, true
		}
		if text, ok := stringField(itemMap, "text"); ok {
			return text, true
		}
		content, ok := itemMap["content"].([]any)
		if !ok {
			continue
		}
		for _, entry := range content {
			entryMap, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			for _, key := range []string{"text", "output_text"} {
				if text, ok := stringField(entryMap, key); ok {
					return text, true
				}
			}
			if raw, ok := entryMap["json"]; ok {
				encoded, err := json.Marshal(raw)
				if err == nil {
					return string(encoded), true
				}
			}
		}
	}
	return "", false
}

func stringField(payload map[string]any, key string) (string, bool) {
	value, ok := payload[key].(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", false
	}
	return value, true
}

func trimJSONFence(text string) string {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "```") {
		return text
	}
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSpace(text)
	if strings.HasPrefix(strings.ToLower(text), "json") {
		text = strings.TrimSpace(text[4:])
	}
	if index := strings.LastIndex(text, "```"); index >= 0 {
		text = text[:index]
	}
	return strings.TrimSpace(text)
}

func extractResultMetadata(raw json.RawMessage) (float64, string) {
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return 0, ""
	}
	var confidence float64
	if value, ok := payload["confidence"].(float64); ok {
		confidence = value
	}
	uncertainty, _ := payload["uncertainty"].(string)
	return confidence, uncertainty
}

func cloneRaw(raw json.RawMessage) json.RawMessage {
	if raw == nil {
		return nil
	}
	cloned := make([]byte, len(raw))
	copy(cloned, raw)
	return cloned
}

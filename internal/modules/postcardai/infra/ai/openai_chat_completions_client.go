package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
)

type OpenAIChatCompletionsClient struct {
	apiKey     string
	model      string
	endpoint   string
	httpClient *http.Client
}

func NewOpenAIChatCompletionsClient(apiKey, model, endpoint string, timeout time.Duration) *OpenAIChatCompletionsClient {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1/chat/completions"
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &OpenAIChatCompletionsClient{
		apiKey:     apiKey,
		model:      model,
		endpoint:   endpoint,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *OpenAIChatCompletionsClient) AnalyzeImage(ctx context.Context, req postcardaiapp.ImageUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	if c.apiKey == "" {
		return nil, domain.ProviderError{Code: domain.ErrorProviderUnavailable, Err: errors.New("missing API key")}
	}
	payload := map[string]any{
		"model": c.model,
		"messages": []map[string]any{{
			"role": "user",
			"content": []map[string]any{
				{"type": "text", "text": ImagePromptV1()},
				{"type": "image_url", "image_url": map[string]string{"url": req.SignedURL}},
			},
		}},
	}
	return c.do(ctx, payload)
}

func (c *OpenAIChatCompletionsClient) RepairImage(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return c.repair(ctx, req)
}

func (c *OpenAIChatCompletionsClient) AnalyzePostcard(ctx context.Context, req postcardaiapp.PostcardUnderstandingRequest) (*postcardaiapp.AIResult, error) {
	if c.apiKey == "" {
		return nil, domain.ProviderError{Code: domain.ErrorProviderUnavailable, Err: errors.New("missing API key")}
	}
	payload := map[string]any{
		"model": c.model,
		"messages": []map[string]string{{
			"role":    "user",
			"content": PostcardPromptV1() + "\n\nPostcard content:\n" + string(req.Postcard.Content),
		}},
	}
	return c.do(ctx, payload)
}

func (c *OpenAIChatCompletionsClient) RepairPostcard(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	return c.repair(ctx, req)
}

func (c *OpenAIChatCompletionsClient) repair(ctx context.Context, req postcardaiapp.RepairRequest) (*postcardaiapp.AIResult, error) {
	if c.apiKey == "" {
		return nil, domain.ProviderError{Code: domain.ErrorProviderUnavailable, Err: errors.New("missing API key")}
	}
	payload := map[string]any{
		"model": c.model,
		"messages": []map[string]string{{
			"role":    "user",
			"content": "Repair this JSON for schema " + req.SchemaVersion + ": " + string(req.Original),
		}},
	}
	return c.do(ctx, payload)
}

func (c *OpenAIChatCompletionsClient) do(ctx context.Context, payload any) (*postcardaiapp.AIResult, error) {
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
	if response.StatusCode == http.StatusForbidden || response.StatusCode == http.StatusBadRequest || response.StatusCode == http.StatusUnauthorized {
		return nil, domain.ProviderError{Code: domain.ErrorProviderRefused, Err: errors.New(response.Status)}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, domain.ProviderError{Code: domain.ErrorProviderRefused, Err: errors.New(response.Status)}
	}
	raw, err := readBounded(response.Body, maxResponseBodyBytes)
	if err != nil {
		return nil, domain.ProviderError{Code: domain.ErrorMalformedOutput, Err: err}
	}
	result, err := parseChatCompletionsResult(raw)
	if err != nil {
		return nil, domain.ProviderError{Code: domain.ErrorMalformedOutput, Err: err}
	}
	return result, nil
}

func parseChatCompletionsResult(raw []byte) (*postcardaiapp.AIResult, error) {
	var payload struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	if len(payload.Choices) == 0 || payload.Choices[0].Message.Content == "" {
		return nil, fmt.Errorf("chat completion did not include message content")
	}
	text := trimJSONFence(payload.Choices[0].Message.Content)
	if !json.Valid([]byte(text)) {
		return nil, errors.New("chat completion content is not valid JSON")
	}
	result := json.RawMessage(text)
	confidence, uncertainty := extractResultMetadata(result)
	return &postcardaiapp.AIResult{
		JSON:        cloneRaw(result),
		Confidence:  confidence,
		Uncertainty: uncertainty,
	}, nil
}

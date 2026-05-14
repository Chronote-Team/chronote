package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
)

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
	return nil, domain.ProviderError{Code: domain.ErrorMalformedOutput, Err: errors.New("response parsing not configured for live provider")}
}

package app

import (
	"context"
	"time"
)

type NoopEnqueuer struct{}

func (NoopEnqueuer) EnqueuePostcardAnalysis(context.Context, EnqueueInput) (*EnqueueResult, error) {
	return &EnqueueResult{Noop: true}, nil
}

type NoopStorage struct{}

func (NoopStorage) PresignGetObject(ctx context.Context, objectKey string, ttl time.Duration) (string, error) {
	return "https://signed.example.com/" + objectKey, nil
}

type NoopAIClient struct{}

func (NoopAIClient) AnalyzeImage(ctx context.Context, req ImageUnderstandingRequest) (*AIResult, error) {
	return &AIResult{JSON: []byte(`{"image_type":"memory","caption":"test image","confidence":0.9}`), Confidence: 0.9}, nil
}

func (NoopAIClient) RepairImage(ctx context.Context, req RepairRequest) (*AIResult, error) {
	return &AIResult{JSON: []byte(`{"image_type":"memory","caption":"repaired image","confidence":0.8}`), Confidence: 0.8}, nil
}

func (NoopAIClient) AnalyzePostcard(ctx context.Context, req PostcardUnderstandingRequest) (*AIResult, error) {
	uncertainty := req.Uncertainty
	if uncertainty == "" && req.Partial {
		uncertainty = "partial media analysis"
	}
	return &AIResult{JSON: []byte(`{"summary":"test postcard memory","suggested_title":"Memory","confidence":0.9}`), Confidence: 0.9, Uncertainty: uncertainty}, nil
}

func (NoopAIClient) RepairPostcard(ctx context.Context, req RepairRequest) (*AIResult, error) {
	return &AIResult{JSON: []byte(`{"summary":"repaired postcard memory","suggested_title":"Memory","confidence":0.8}`), Confidence: 0.8}, nil
}

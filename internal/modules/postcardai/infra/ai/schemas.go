package ai

import (
	"encoding/json"
	"errors"
)

func ValidateImageUnderstanding(raw json.RawMessage) error {
	var payload struct {
		ImageType  string   `json:"image_type"`
		Caption    string   `json:"caption"`
		Confidence *float64 `json:"confidence"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	if payload.ImageType == "" {
		return errors.New("image_type is required")
	}
	if payload.Caption == "" {
		return errors.New("caption is required")
	}
	if payload.Confidence == nil {
		return errors.New("confidence is required")
	}
	return nil
}

func ValidatePostcardUnderstanding(raw json.RawMessage) error {
	var payload struct {
		Summary        string   `json:"summary"`
		SuggestedTitle string   `json:"suggested_title"`
		Confidence     *float64 `json:"confidence"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return err
	}
	if payload.Summary == "" {
		return errors.New("summary is required")
	}
	if payload.SuggestedTitle == "" {
		return errors.New("suggested_title is required")
	}
	if payload.Confidence == nil {
		return errors.New("confidence is required")
	}
	return nil
}

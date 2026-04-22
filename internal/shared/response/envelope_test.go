package response

import (
	"encoding/json"
	"testing"
)

func TestSuccessEnvelopeIncludesCodeMessageAndData(t *testing.T) {
	body, err := JSON(200, "OK", map[string]string{"status": "up"})
	if err != nil {
		t.Fatalf("JSON returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if payload["code"] != float64(200) {
		t.Fatalf("expected code 200, got %#v", payload["code"])
	}
	if payload["message"] != "OK" {
		t.Fatalf("expected OK message, got %#v", payload["message"])
	}
	if payload["data"] == nil {
		t.Fatalf("expected data field to be present")
	}
}

func TestErrorEnvelopeOmitsDataWhenNil(t *testing.T) {
	body, err := JSON(401, "Without Token, and you're unauthorized!", nil)
	if err != nil {
		t.Fatalf("JSON returned error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if _, exists := payload["data"]; exists {
		t.Fatalf("expected no data field for nil payload")
	}
}

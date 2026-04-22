package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appplatform "chronote-refactor/internal/platform/app"
)

func TestHealthContractMatchesEnvelope(t *testing.T) {
	app, err := appplatform.NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	app.Router().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", recorder.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if payload["code"] != float64(http.StatusServiceUnavailable) {
		t.Fatalf("unexpected code: %#v", payload["code"])
	}
	if payload["message"] != "Service Unavailable" {
		t.Fatalf("unexpected message: %#v", payload["message"])
	}
}

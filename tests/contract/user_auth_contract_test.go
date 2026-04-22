package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	appplatform "chronote-refactor/internal/platform/app"
)

func TestRegisterContractMatchesEnvelope(t *testing.T) {
	app, err := appplatform.NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp returned error: %v", err)
	}

	body := bytes.NewBufferString(`{"username":"tester","email":"tester@example.com","password":"123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/user/register", body)
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	app.Router().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", recorder.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload["message"] != "用户注册成功" {
		t.Fatalf("unexpected message: %#v", payload["message"])
	}
}

package contract

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
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

func TestAvatarUploadPersistsUsableStorageURL(t *testing.T) {
	app, err := appplatform.NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp returned error: %v", err)
	}
	accessToken := contractRegisterAndLogin(t, app, "avatar_user", "avatar@example.com")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="avatar"; filename="avatar.png"`)
	header.Set("Content-Type", "image/png")
	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("create avatar file: %v", err)
	}
	if _, err := part.Write([]byte("fake-png-data")); err != nil {
		t.Fatalf("write avatar data: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/user/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+accessToken)
	recorder := httptest.NewRecorder()
	app.Router().ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected avatar upload 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	payload := decodeContractPayload(t, recorder.Body.Bytes())
	if payload["message"] != "头像上传成功" {
		t.Fatalf("unexpected avatar upload message: %#v", payload["message"])
	}
	data := payload["data"].(map[string]any)
	avatarURL := data["avatar_url"].(string)
	if avatarURL == "" || avatarURL[0] == '/' || avatarURL == "https://cdn.example.com/avatars/1/avatar.png" {
		t.Fatalf("expected usable storage avatar URL, got %q", avatarURL)
	}

	infoReq := httptest.NewRequest(http.MethodGet, "/user/info", nil)
	infoReq.Header.Set("Authorization", "Bearer "+accessToken)
	infoRecorder := httptest.NewRecorder()
	app.Router().ServeHTTP(infoRecorder, infoReq)
	if infoRecorder.Code != http.StatusOK {
		t.Fatalf("expected user info 200, got %d body=%s", infoRecorder.Code, infoRecorder.Body.String())
	}
	infoPayload := decodeContractPayload(t, infoRecorder.Body.Bytes())
	infoData := infoPayload["data"].(map[string]any)
	if infoData["avatar"] != avatarURL {
		t.Fatalf("expected user info avatar %q, got %#v", avatarURL, infoData["avatar"])
	}
}

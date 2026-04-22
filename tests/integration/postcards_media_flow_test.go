package integration

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostcardsMediaFlow(t *testing.T) {
	app, err := NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp returned error: %v", err)
	}

	accessToken := integrationRegisterAndLogin(t, app)
	postcardID := integrationCreatePostcard(t, app, accessToken)
	mediaID := integrationUploadMedia(t, app, accessToken, postcardID)

	deleteMediaReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/postcards/%d/media/%d", postcardID, mediaID), nil)
	deleteMediaReq.Header.Set("Authorization", "Bearer "+accessToken)
	deleteMediaRes := httptest.NewRecorder()
	app.Router().ServeHTTP(deleteMediaRes, deleteMediaReq)

	if deleteMediaRes.Code != http.StatusOK {
		t.Fatalf("expected delete media 200, got %d body=%s", deleteMediaRes.Code, deleteMediaRes.Body.String())
	}

	deletePostcardReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/postcards/%d", postcardID), nil)
	deletePostcardReq.Header.Set("Authorization", "Bearer "+accessToken)
	deletePostcardRes := httptest.NewRecorder()
	app.Router().ServeHTTP(deletePostcardRes, deletePostcardReq)

	if deletePostcardRes.Code != http.StatusOK {
		t.Fatalf("expected delete postcard 200, got %d body=%s", deletePostcardRes.Code, deletePostcardRes.Body.String())
	}
}

func integrationRegisterAndLogin(t *testing.T, app *TestApp) string {
	t.Helper()

	registerBody := bytes.NewBufferString(`{"username":"flow","email":"flow@example.com","password":"123456"}`)
	registerReq := httptest.NewRequest(http.MethodPost, "/user/register", registerBody)
	registerReq.Header.Set("Content-Type", "application/json")
	registerRes := httptest.NewRecorder()
	app.Router().ServeHTTP(registerRes, registerReq)

	if registerRes.Code != http.StatusCreated {
		t.Fatalf("expected register 201, got %d body=%s", registerRes.Code, registerRes.Body.String())
	}

	loginBody := bytes.NewBufferString(`{"email":"flow@example.com","password":"123456"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/user/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/json")
	loginRes := httptest.NewRecorder()
	app.Router().ServeHTTP(loginRes, loginReq)

	if loginRes.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d body=%s", loginRes.Code, loginRes.Body.String())
	}

	return extractAccessToken(t, loginRes.Body.Bytes())
}

func integrationCreatePostcard(t *testing.T, app *TestApp, accessToken string) uint {
	t.Helper()

	body := bytes.NewBufferString(`{"title":"Flow Card","content":{"blocks":[{"type":"text","value":"hello"}]},"visibility":"public"}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/postcards", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res := httptest.NewRecorder()
	app.Router().ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected create postcard 201, got %d body=%s", res.Code, res.Body.String())
	}

	return extractObjectID(t, res.Body.Bytes())
}

func integrationUploadMedia(t *testing.T, app *TestApp, accessToken string, postcardID uint) uint {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("media", "cover.jpg")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte("fake-image-data")); err != nil {
		t.Fatalf("write file data: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/v1/postcards/%d/media", postcardID), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res := httptest.NewRecorder()
	app.Router().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected upload media 200, got %d body=%s", res.Code, res.Body.String())
	}

	return extractNestedMediaID(t, res.Body.Bytes())
}

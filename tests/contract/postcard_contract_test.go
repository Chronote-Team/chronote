package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	appplatform "chronote-refactor/internal/platform/app"
)

func TestPostcardContractPreservesVisibilityAndEnvelope(t *testing.T) {
	app, err := appplatform.NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp returned error: %v", err)
	}

	accessToken := contractRegisterAndLogin(t, app, "writer", "writer@example.com")
	publicID := contractCreatePostcard(t, app, accessToken, "Public Card", "public")
	privateID := contractCreatePostcard(t, app, accessToken, "Private Card", "private")

	publicReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/postcards/%d", publicID), nil)
	publicRes := httptest.NewRecorder()
	app.Router().ServeHTTP(publicRes, publicReq)

	if publicRes.Code != http.StatusOK {
		t.Fatalf("expected public detail 200, got %d body=%s", publicRes.Code, publicRes.Body.String())
	}

	publicPayload := decodeContractPayload(t, publicRes.Body.Bytes())
	if publicPayload["message"] != "获取明信片详情成功" {
		t.Fatalf("unexpected public detail message: %#v", publicPayload["message"])
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/postcards", nil)
	listRes := httptest.NewRecorder()
	app.Router().ServeHTTP(listRes, listReq)

	if listRes.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d body=%s", listRes.Code, listRes.Body.String())
	}

	listPayload := decodeContractPayload(t, listRes.Body.Bytes())
	if listPayload["message"] != "获取明信片列表成功" {
		t.Fatalf("unexpected list message: %#v", listPayload["message"])
	}

	data, ok := listPayload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected list data object, got %#v", listPayload["data"])
	}
	items, ok := data["items"].([]any)
	if !ok {
		t.Fatalf("expected list items array, got %#v", data["items"])
	}
	if len(items) != 1 {
		t.Fatalf("expected only public postcard in anonymous list, got %d", len(items))
	}

	privateReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/postcards/%d", privateID), nil)
	privateRes := httptest.NewRecorder()
	app.Router().ServeHTTP(privateRes, privateReq)

	if privateRes.Code != http.StatusForbidden {
		t.Fatalf("expected private detail 403, got %d body=%s", privateRes.Code, privateRes.Body.String())
	}

	privatePayload := decodeContractPayload(t, privateRes.Body.Bytes())
	if privatePayload["message"] != "无权限访问该明信片" {
		t.Fatalf("unexpected private detail message: %#v", privatePayload["message"])
	}
}

func TestRandomPostcardContractReturnsPublicPostcardForAnonymousCaller(t *testing.T) {
	app, err := appplatform.NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp returned error: %v", err)
	}

	accessToken := contractRegisterAndLogin(t, app, "randomwriter", "randomwriter@example.com")
	contractCreatePostcard(t, app, accessToken, "Private Card", "private")
	contractCreatePostcard(t, app, accessToken, "Public Card", "public")

	req := httptest.NewRequest(http.MethodGet, "/v1/postcards/random", nil)
	res := httptest.NewRecorder()
	app.Router().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected random postcard 200, got %d body=%s", res.Code, res.Body.String())
	}

	payload := decodeContractPayload(t, res.Body.Bytes())
	if payload["message"] != "获取随机明信片成功" {
		t.Fatalf("unexpected random postcard message: %#v", payload["message"])
	}
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected random postcard data object, got %#v", payload["data"])
	}
	if data["visibility"] != "public" {
		t.Fatalf("expected anonymous random postcard to be public, got %#v", data["visibility"])
	}
}

func TestRandomPostcardContractIncludesOwnedPrivatePostcardForAuthenticatedCaller(t *testing.T) {
	app, err := appplatform.NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp returned error: %v", err)
	}

	otherToken := contractRegisterAndLogin(t, app, "otherwriter", "otherwriter@example.com")
	contractCreatePostcard(t, app, otherToken, "Other Private Card", "private")
	ownerToken := contractRegisterAndLogin(t, app, "ownerwriter", "ownerwriter@example.com")
	ownedID := contractCreatePostcard(t, app, ownerToken, "Owned Private Card", "private")

	req := httptest.NewRequest(http.MethodGet, "/v1/postcards/random", nil)
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	res := httptest.NewRecorder()
	app.Router().ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected random postcard 200, got %d body=%s", res.Code, res.Body.String())
	}

	payload := decodeContractPayload(t, res.Body.Bytes())
	if payload["message"] != "获取随机明信片成功" {
		t.Fatalf("unexpected random postcard message: %#v", payload["message"])
	}
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected random postcard data object, got %#v", payload["data"])
	}
	if gotID := uint(data["id"].(float64)); gotID != ownedID {
		t.Fatalf("expected owned private postcard %d, got %d", ownedID, gotID)
	}
	if data["visibility"] != "private" {
		t.Fatalf("expected owned private visibility, got %#v", data["visibility"])
	}
}

func TestRandomPostcardContractReturnsNotFoundWhenNoPostcardsAreAccessible(t *testing.T) {
	app, err := appplatform.NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp returned error: %v", err)
	}

	accessToken := contractRegisterAndLogin(t, app, "privateonly", "privateonly@example.com")
	contractCreatePostcard(t, app, accessToken, "Private Card", "private")

	req := httptest.NewRequest(http.MethodGet, "/v1/postcards/random", nil)
	res := httptest.NewRecorder()
	app.Router().ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected random postcard 404, got %d body=%s", res.Code, res.Body.String())
	}

	payload := decodeContractPayload(t, res.Body.Bytes())
	if payload["message"] != "明信片不存在" {
		t.Fatalf("unexpected not found message: %#v", payload["message"])
	}
	if payload["data"] != nil {
		t.Fatalf("expected null data, got %#v", payload["data"])
	}
}

func TestRandomPostcardContractUsesRandomRouteInsteadOfDetailRoute(t *testing.T) {
	app, err := appplatform.NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/postcards/random", nil)
	res := httptest.NewRecorder()
	app.Router().ServeHTTP(res, req)

	if res.Code == http.StatusBadRequest {
		t.Fatalf("expected random route handling, got detail route bad request body=%s", res.Body.String())
	}
	payload := decodeContractPayload(t, res.Body.Bytes())
	if payload["message"] == "明信片 ID 无效" {
		t.Fatalf("expected random route handling, got detail route message")
	}
}

func contractRegisterAndLogin(t *testing.T, app *appplatform.App, username, email string) string {
	t.Helper()

	registerBody := bytes.NewBufferString(fmt.Sprintf(`{"username":"%s","email":"%s","password":"123456"}`, username, email))
	registerReq := httptest.NewRequest(http.MethodPost, "/user/register", registerBody)
	registerReq.Header.Set("Content-Type", "application/json")
	registerRes := httptest.NewRecorder()
	app.Router().ServeHTTP(registerRes, registerReq)

	if registerRes.Code != http.StatusCreated {
		t.Fatalf("expected register 201, got %d body=%s", registerRes.Code, registerRes.Body.String())
	}

	loginBody := bytes.NewBufferString(fmt.Sprintf(`{"email":"%s","password":"123456"}`, email))
	loginReq := httptest.NewRequest(http.MethodPost, "/user/login", loginBody)
	loginReq.Header.Set("Content-Type", "application/json")
	loginRes := httptest.NewRecorder()
	app.Router().ServeHTTP(loginRes, loginReq)

	if loginRes.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d body=%s", loginRes.Code, loginRes.Body.String())
	}

	payload := decodeContractPayload(t, loginRes.Body.Bytes())
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected login data object, got %#v", payload["data"])
	}
	accessToken, ok := data["access_token"].(string)
	if !ok || accessToken == "" {
		t.Fatalf("expected access token, got %#v", data["access_token"])
	}
	return accessToken
}

func contractCreatePostcard(t *testing.T, app *appplatform.App, accessToken, title, visibility string) uint {
	t.Helper()

	body := bytes.NewBufferString(fmt.Sprintf(`{"title":"%s","content":{"blocks":[{"type":"text","value":"hello"}]},"visibility":"%s"}`, title, visibility))
	req := httptest.NewRequest(http.MethodPost, "/v1/postcards", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res := httptest.NewRecorder()
	app.Router().ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected create postcard 201, got %d body=%s", res.Code, res.Body.String())
	}

	payload := decodeContractPayload(t, res.Body.Bytes())
	if payload["message"] != "明信片创建成功" {
		t.Fatalf("unexpected create postcard message: %#v", payload["message"])
	}

	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected postcard data object, got %#v", payload["data"])
	}
	id, ok := data["id"].(float64)
	if !ok || id == 0 {
		t.Fatalf("expected postcard id, got %#v", data["id"])
	}
	return uint(id)
}

func decodeContractPayload(t *testing.T, body []byte) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	return payload
}

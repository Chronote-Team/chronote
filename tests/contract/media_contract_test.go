package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	appplatform "chronote-refactor/internal/platform/app"
)

func TestMediaContractPreservesUploadListAndReorderRules(t *testing.T) {
	app, err := appplatform.NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp returned error: %v", err)
	}

	accessToken := contractRegisterAndLogin(t, app, "artist", "artist@example.com")
	postcardID := contractCreatePostcard(t, app, accessToken, "Media Card", "public")

	uploadBody := &bytes.Buffer{}
	writer := multipart.NewWriter(uploadBody)
	if err := writer.WriteField("media_group", "gallery"); err != nil {
		t.Fatalf("write media_group: %v", err)
	}
	for _, name := range []string{"a.jpg", "b.jpg"} {
		part, err := writer.CreateFormFile("medias", name)
		if err != nil {
			t.Fatalf("create form file: %v", err)
		}
		if _, err := part.Write([]byte("fake-image-data")); err != nil {
			t.Fatalf("write file data: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	uploadReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/v1/postcards/%d/media", postcardID), uploadBody)
	uploadReq.Header.Set("Content-Type", writer.FormDataContentType())
	uploadReq.Header.Set("Authorization", "Bearer "+accessToken)
	uploadRes := httptest.NewRecorder()
	app.Router().ServeHTTP(uploadRes, uploadReq)

	if uploadRes.Code != http.StatusOK {
		t.Fatalf("expected upload 200, got %d body=%s", uploadRes.Code, uploadRes.Body.String())
	}

	uploadPayload := decodeContractPayload(t, uploadRes.Body.Bytes())
	if uploadPayload["message"] != "媒体上传成功" {
		t.Fatalf("unexpected upload message: %#v", uploadPayload["message"])
	}
	uploadData := uploadPayload["data"].(map[string]any)
	medias := uploadData["medias"].([]any)
	if len(medias) != 2 {
		t.Fatalf("expected two uploaded medias, got %d", len(medias))
	}

	listReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/postcards/%d/media", postcardID), nil)
	listRes := httptest.NewRecorder()
	app.Router().ServeHTTP(listRes, listReq)

	if listRes.Code != http.StatusOK {
		t.Fatalf("expected media list 200, got %d body=%s", listRes.Code, listRes.Body.String())
	}

	listPayload := decodeContractPayload(t, listRes.Body.Bytes())
	if listPayload["message"] != "获取媒体列表成功" {
		t.Fatalf("unexpected media list message: %#v", listPayload["message"])
	}

	firstID := medias[0].(map[string]any)["id"]
	secondID := medias[1].(map[string]any)["id"]

	invalidReorderBody := bytes.NewBufferString(fmt.Sprintf(`{"media_ids":[%.0f]}`, firstID))
	invalidReorderReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/postcards/%d/media/reorder", postcardID), invalidReorderBody)
	invalidReorderReq.Header.Set("Content-Type", "application/json")
	invalidReorderReq.Header.Set("Authorization", "Bearer "+accessToken)
	invalidReorderRes := httptest.NewRecorder()
	app.Router().ServeHTTP(invalidReorderRes, invalidReorderReq)

	if invalidReorderRes.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid reorder 400, got %d body=%s", invalidReorderRes.Code, invalidReorderRes.Body.String())
	}

	invalidPayload := decodeContractPayload(t, invalidReorderRes.Body.Bytes())
	if invalidPayload["message"] != "必须传入全部媒体ID进行排序" {
		t.Fatalf("unexpected invalid reorder message: %#v", invalidPayload["message"])
	}

	reorderBody := bytes.NewBufferString(fmt.Sprintf(`{"media_ids":[%.0f,%.0f]}`, secondID, firstID))
	reorderReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/postcards/%d/media/reorder", postcardID), reorderBody)
	reorderReq.Header.Set("Content-Type", "application/json")
	reorderReq.Header.Set("Authorization", "Bearer "+accessToken)
	reorderRes := httptest.NewRecorder()
	app.Router().ServeHTTP(reorderRes, reorderReq)

	if reorderRes.Code != http.StatusOK {
		t.Fatalf("expected reorder 200, got %d body=%s", reorderRes.Code, reorderRes.Body.String())
	}

	reorderPayload := decodeContractPayload(t, reorderRes.Body.Bytes())
	if reorderPayload["message"] != "媒体排序更新成功" {
		t.Fatalf("unexpected reorder message: %#v", reorderPayload["message"])
	}
}

func decodeMediaPayload(t *testing.T, body []byte) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	return payload
}

package integration

import "encoding/json"

func extractAccessToken(t testingT, body []byte) string {
	payload := decodePayload(t, body)
	data := payload["data"].(map[string]any)
	return data["access_token"].(string)
}

func extractObjectID(t testingT, body []byte) uint {
	payload := decodePayload(t, body)
	data := payload["data"].(map[string]any)
	return uint(data["id"].(float64))
}

func extractNestedMediaID(t testingT, body []byte) uint {
	payload := decodePayload(t, body)
	data := payload["data"].(map[string]any)
	medias := data["medias"].([]any)
	return uint(medias[0].(map[string]any)["id"].(float64))
}

type testingT interface {
	Helper()
	Fatalf(format string, args ...any)
}

func decodePayload(t testingT, body []byte) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	return payload
}

package contract

import (
	"net/http"
	"net/http/httptest"
	"testing"

	appplatform "chronote-refactor/internal/platform/app"
)

func TestAnalysisEndpointsAreNotClientFacing(t *testing.T) {
	app, err := appplatform.NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp returned error: %v", err)
	}

	for _, tc := range []struct {
		method string
		path   string
	}{
		{method: http.MethodPost, path: "/v1/postcards/1/analyze"},
		{method: http.MethodGet, path: "/v1/postcards/1/analysis"},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		res := httptest.NewRecorder()
		app.Router().ServeHTTP(res, req)
		if res.Code != http.StatusNotFound {
			t.Fatalf("%s %s: expected 404 for absent analysis endpoint, got %d body=%s", tc.method, tc.path, res.Code, res.Body.String())
		}
	}
}

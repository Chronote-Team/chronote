package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	appplatform "chronote-refactor/internal/platform/app"
)

func assertCutoverCompatibility(t *testing.T, app *appplatform.App, fixture CutoverFixture) {
	t.Helper()

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	healthRec := httptest.NewRecorder()
	app.Router().ServeHTTP(healthRec, healthReq)
	if healthRec.Code != http.StatusOK {
		t.Fatalf("expected /health 200, got %d body=%s", healthRec.Code, healthRec.Body.String())
	}

	publicReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/postcards/%d", fixture.PublicPostcardID), nil)
	publicRec := httptest.NewRecorder()
	app.Router().ServeHTTP(publicRec, publicReq)
	if publicRec.Code != http.StatusOK {
		t.Fatalf("expected public postcard detail 200, got %d body=%s", publicRec.Code, publicRec.Body.String())
	}

	privateReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/postcards/%d", fixture.PrivatePostcardID), nil)
	privateRec := httptest.NewRecorder()
	app.Router().ServeHTTP(privateRec, privateReq)
	if privateRec.Code != http.StatusForbidden {
		t.Fatalf("expected private postcard detail 403, got %d body=%s", privateRec.Code, privateRec.Body.String())
	}

	mediaReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/postcards/%d/media", fixture.PublicPostcardID), nil)
	mediaRec := httptest.NewRecorder()
	app.Router().ServeHTTP(mediaRec, mediaReq)
	if mediaRec.Code != http.StatusOK {
		t.Fatalf("expected public media list 200, got %d body=%s", mediaRec.Code, mediaRec.Body.String())
	}
}

func assertDegradedHealth(t *testing.T, app *appplatform.App) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/health/details", nil)
	rec := httptest.NewRecorder()
	app.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusMultiStatus {
		t.Fatalf("expected /health/details 207 under degraded dependency, got %d body=%s", rec.Code, rec.Body.String())
	}
}

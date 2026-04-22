package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFullStackVerificationMatrix(t *testing.T) {
	env := requireCutoverEnv(t)
	fixture := loadSupportedFixture(t, env)
	app := newProductionApp(t)

	t.Run("health and supported data", func(t *testing.T) {
		assertCutoverCompatibility(t, app, fixture)
	})

	t.Run("auth flow uses external dependencies", func(t *testing.T) {
		registerBody := bytes.NewBufferString(`{"username":"cutover_register","email":"cutover-register@example.com","password":"123456"}`)
		registerReq := httptest.NewRequest(http.MethodPost, "/user/register", registerBody)
		registerReq.Header.Set("Content-Type", "application/json")
		registerRec := httptest.NewRecorder()
		app.Router().ServeHTTP(registerRec, registerReq)
		if registerRec.Code != http.StatusCreated {
			t.Fatalf("expected register 201, got %d body=%s", registerRec.Code, registerRec.Body.String())
		}

		token := loginFixtureUser(t, app)
		infoReq := httptest.NewRequest(http.MethodGet, "/user/info", nil)
		infoReq.Header.Set("Authorization", "Bearer "+token)
		infoRec := httptest.NewRecorder()
		app.Router().ServeHTTP(infoRec, infoReq)
		if infoRec.Code != http.StatusOK {
			t.Fatalf("expected user info 200, got %d body=%s", infoRec.Code, infoRec.Body.String())
		}
	})
}

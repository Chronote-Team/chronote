package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

func TestDockerfileBuildsAPIAndWorkerBinaries(t *testing.T) {
	content, err := os.ReadFile("../../Dockerfile")
	if err != nil {
		t.Fatalf("read Dockerfile: %v", err)
	}
	text := string(content)
	for _, expected := range []string{
		"go build -o /out/chronote ./cmd/api",
		"go build -o /out/chronote-worker ./cmd/worker",
		"COPY --from=builder /out/chronote /app/chronote",
		"COPY --from=builder /out/chronote-worker /app/chronote-worker",
		`ENTRYPOINT ["/app/chronote"]`,
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("Dockerfile missing %q", expected)
		}
	}
}

func TestComposeDefinesWorkerServiceWithoutPorts(t *testing.T) {
	content, err := os.ReadFile("../../docker-compose.yml")
	if err != nil {
		t.Fatalf("read docker-compose.yml: %v", err)
	}
	text := string(content)
	for _, expected := range []string{
		"worker:",
		"container_name: chronote-worker",
		`entrypoint: ["/app/chronote-worker"]`,
		"AI_WORKER_ID:",
		"AI_WORKER_IDLE_SLEEP:",
		"AI_WORKER_ERROR_SLEEP:",
		"AI_WORKER_RUN_ONCE:",
		"condition: service_completed_successfully",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("docker-compose.yml missing %q", expected)
		}
	}
	workerSection := text[strings.Index(text, "  worker:"):]
	if strings.Contains(workerSection, "\n    ports:") {
		t.Fatalf("worker service must not expose ports")
	}
}

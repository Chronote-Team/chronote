package integration

import "testing"

func TestCutoverCompatibilityAgainstSupportedData(t *testing.T) {
	env := requireCutoverEnv(t)
	fixture := loadSupportedFixture(t, env)
	app := newProductionApp(t)

	assertCutoverCompatibility(t, app, fixture)
}

func TestHealthDetailsDegradesWhenRedisUnavailable(t *testing.T) {
	_ = requireCutoverEnv(t)

	t.Setenv("REDIS_HOST", "127.0.0.1")
	t.Setenv("REDIS_PORT", "6399")

	app := newProductionApp(t)
	assertDegradedHealth(t, app)
}

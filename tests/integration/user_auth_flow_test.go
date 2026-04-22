package integration

import "testing"

func TestUserAuthFlow(t *testing.T) {
	app, err := NewTestApp()
	if err != nil {
		t.Fatalf("NewTestApp returned error: %v", err)
	}
	if app == nil {
		t.Fatalf("expected test app")
	}
}

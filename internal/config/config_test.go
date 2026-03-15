package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	_ = os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	_ = os.Setenv("TG_SESSION_STRING", "session_string")
	_ = os.Setenv("TG_API_ID", "12345")
	_ = os.Setenv("TG_API_HASH", "hash")
	_ = os.Setenv("PINPOINT_AUTH_TOKEN", "token")

	defer func() {
		_ = os.Unsetenv("MONGODB_URI")
		_ = os.Unsetenv("TG_SESSION_STRING")
		_ = os.Unsetenv("TG_API_ID")
		_ = os.Unsetenv("TG_API_HASH")
		_ = os.Unsetenv("PINPOINT_AUTH_TOKEN")
	}()

	cfg := Load()

	if cfg.MongoURI != "mongodb://localhost:27017" {
		t.Errorf("expected MONGODB_URI to be mongodb://localhost:27017, got %s", cfg.MongoURI)
	}
	if cfg.PinpointAuthToken != "token" {
		t.Errorf("expected PinpointAuthToken to be token, got %s", cfg.PinpointAuthToken)
	}
	if cfg.TGAPIID != 12345 {
		t.Errorf("expected TG_API_ID to be 12345, got %d", cfg.TGAPIID)
	}
}

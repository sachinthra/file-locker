package config

import (
	"testing"
)

func TestLoadConfig(t *testing.T) {
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test that values were loaded correctly
	if config.Server.Port != 9010 {
		t.Errorf("Expected server port 9010, got %d", config.Server.Port)
	}

	if config.Server.GRPCPort != 9011 {
		t.Errorf("Expected gRPC port 9011, got %d", config.Server.GRPCPort)
	}

	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected host 0.0.0.0, got %s", config.Server.Host)
	}

	if config.Storage.MinIO.Endpoint != "localhost:9012" {
		t.Errorf("Expected MinIO endpoint localhost:9012, got %s", config.Storage.MinIO.Endpoint)
	}

	if config.Storage.MinIO.Bucket != "filelocker" {
		t.Errorf("Expected MinIO bucket filelocker, got %s", config.Storage.MinIO.Bucket)
	}

	if config.Storage.Redis.Addr != "localhost:6379" {
		t.Errorf("Expected Redis addr localhost:6379, got %s", config.Storage.Redis.Addr)
	}

	if config.Security.JWTSecret != "change-me-in-production" {
		t.Errorf("Expected JWT secret, got %s", config.Security.JWTSecret)
	}

	if !config.Features.AutoDelete.Enabled {
		t.Error("Expected auto_delete to be enabled")
	}

	if config.Features.AutoDelete.CheckInterval != 3600 {
		t.Errorf("Expected check_interval 3600, got %d", config.Features.AutoDelete.CheckInterval)
	}

	if config.Logging.Level != "info" {
		t.Errorf("Expected log level info, got %s", config.Logging.Level)
	}

	t.Logf("Config loaded successfully: %+v", config)
}

package config

import "testing"

const testDatabaseURL = "postgres://sentinelops:test@localhost:5432/sentinelops"

func TestLoadUsesDefaultPort(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", testDatabaseURL)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Address != ":8080" {
		t.Errorf(
			"expected address %q, got %q",
			":8080",
			cfg.Address,
		)
	}
}

func TestLoadUsesEnvironmentPort(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", testDatabaseURL)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Address != ":9090" {
		t.Errorf(
			"expected address %q, got %q",
			":9090",
			cfg.Address,
		)
	}
}

func TestLoadRejectsNonNumberPort(t *testing.T) {

	t.Setenv("PORT", "banana")
	t.Setenv("DATABSE_URL", testDatabaseURL)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestLoadRejectsOutOfRangePort(t *testing.T) {
	t.Setenv("PORT", "700000")
	t.Setenv("DATABASE_URL", testDatabaseURL)

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func testLoadRejectsMissingDatabaseURL(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error, got nil")

	}
}

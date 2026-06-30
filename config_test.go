package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// writeConfig writes a config.yaml containing key under
// $XDG_CONFIG_HOME/dmrfill/ and returns its path.
func writeConfig(t *testing.T, xdgDir, key string) string {
	t.Helper()
	dir := filepath.Join(xdgDir, "dmrfill")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("repeaterbook_api_key: "+key+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile config: %v", err)
	}
	return path
}

// writeKeyFile writes key to a file in the test temp dir and returns its path.
func writeKeyFile(t *testing.T, key string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "rb.key")
	if err := os.WriteFile(path, []byte("  "+key+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile key: %v", err)
	}
	return path
}

// TestRepeaterBookKeyPrecedence verifies the resolution order:
// -rb-key-file > REPEATERBOOK_API_KEY > config file > error.
func TestRepeaterBookKeyPrecedence(t *testing.T) {
	t.Run("file wins over env and config", func(t *testing.T) {
		xdg := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", xdg)
		t.Setenv(repeaterBookKeyEnv, "env-key")
		writeConfig(t, xdg, "config-key")
		keyFile := writeKeyFile(t, "file-key")

		got, err := repeaterBookKey(keyFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "file-key" {
			t.Errorf("got %q, want %q", got, "file-key")
		}
	})

	t.Run("env wins over config", func(t *testing.T) {
		xdg := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", xdg)
		t.Setenv(repeaterBookKeyEnv, "env-key")
		writeConfig(t, xdg, "config-key")

		got, err := repeaterBookKey("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "env-key" {
			t.Errorf("got %q, want %q", got, "env-key")
		}
	})

	t.Run("config used when no file or env", func(t *testing.T) {
		xdg := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", xdg)
		t.Setenv(repeaterBookKeyEnv, "")
		writeConfig(t, xdg, "config-key")

		got, err := repeaterBookKey("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "config-key" {
			t.Errorf("got %q, want %q", got, "config-key")
		}
	})

	t.Run("noKeyError when nothing is set", func(t *testing.T) {
		xdg := t.TempDir() // empty: no config file present
		t.Setenv("XDG_CONFIG_HOME", xdg)
		t.Setenv(repeaterBookKeyEnv, "")

		_, err := repeaterBookKey("")
		var nke *noKeyError
		if !errors.As(err, &nke) {
			t.Fatalf("got error %v, want *noKeyError", err)
		}
	})

	t.Run("empty env falls through to config", func(t *testing.T) {
		xdg := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", xdg)
		t.Setenv(repeaterBookKeyEnv, "   ") // whitespace only
		writeConfig(t, xdg, "config-key")

		got, err := repeaterBookKey("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "config-key" {
			t.Errorf("got %q, want %q", got, "config-key")
		}
	})

	t.Run("missing key file is an error, not fallthrough", func(t *testing.T) {
		xdg := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", xdg)
		t.Setenv(repeaterBookKeyEnv, "env-key")

		_, err := repeaterBookKey(filepath.Join(t.TempDir(), "does-not-exist"))
		if err == nil {
			t.Fatal("expected error for missing key file, got nil")
		}
		var nke *noKeyError
		if errors.As(err, &nke) {
			t.Fatalf("expected a read error, got noKeyError: %v", err)
		}
	})
}

// TestRedactKey verifies the key is scrubbed from logged strings.
func TestRedactKey(t *testing.T) {
	old := repeaterBookAPIKey
	t.Cleanup(func() { repeaterBookAPIKey = old })

	repeaterBookAPIKey = ""
	if got := redactKey("Authorization: secret"); got != "Authorization: secret" {
		t.Errorf("with no key set, got %q", got)
	}

	repeaterBookAPIKey = "s3cr3t"
	if got := redactKey("Authorization:[s3cr3t]"); got != "Authorization:[[REDACTED]]" {
		t.Errorf("got %q, want redacted", got)
	}
}

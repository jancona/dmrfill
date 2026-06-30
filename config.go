package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const repeaterBookKeyEnv = "REPEATERBOOK_API_KEY"

// repeaterBookAPIKey holds the resolved RepeaterBook API key. It is only
// populated when a datasource that queries RepeaterBook is active. It is a
// secret and must never be logged; use redactKey when logging request details.
var repeaterBookAPIKey string

// config mirrors the dmrfill config file ($XDG_CONFIG_HOME/dmrfill/config.yaml).
type config struct {
	RepeaterBookAPIKey string `yaml:"repeaterbook_api_key"`
}

// noKeyError is returned by repeaterBookKey when no key could be found in any
// of the supported locations. It carries the config file path so callers can
// print a helpful message, and lets RADIOID_DMR distinguish "no key" (fine)
// from a real error reading an explicit key file.
type noKeyError struct {
	configFile string
}

func (e *noKeyError) Error() string {
	return fmt.Sprintf(`no RepeaterBook API key found, but one is required to query RepeaterBook.
Create a key at https://www.repeaterbook.com/user/api_apps.php, then provide it in one of these ways:
  1. -rb-key-file <path>          read the key from a file
  2. export %s=<key>   set it in your shell (add it to your shell rc; note that
     typing it at the prompt records it in your shell history)
  3. repeaterbook_api_key: <key>  add this line to %s (then chmod 600 the file)`,
		repeaterBookKeyEnv, e.configFile)
}

// configPath returns the path to the dmrfill config file, honoring
// XDG_CONFIG_HOME with a fallback to ~/.config.
func configPath() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "dmrfill", "config.yaml"), nil
}

// repeaterBookKey resolves the RepeaterBook API key using this precedence:
//  1. keyFile (the -rb-key-file flag), read and TrimSpace'd
//  2. the REPEATERBOOK_API_KEY environment variable
//  3. repeaterbook_api_key in the config file
//
// If none of these yield a key it returns a *noKeyError describing the options.
// The key never comes from stdin, which is reserved for the codeplug pipeline.
func repeaterBookKey(keyFile string) (string, error) {
	// 1. Explicit key file.
	if keyFile != "" {
		data, err := os.ReadFile(keyFile)
		if err != nil {
			return "", fmt.Errorf("reading RepeaterBook key file %q: %w", keyFile, err)
		}
		key := strings.TrimSpace(string(data))
		if key == "" {
			return "", fmt.Errorf("RepeaterBook key file %q is empty", keyFile)
		}
		return key, nil
	}

	// 2. Environment variable.
	if key := strings.TrimSpace(os.Getenv(repeaterBookKeyEnv)); key != "" {
		return key, nil
	}

	// 3. Config file.
	path, err := configPath()
	if err != nil {
		return "", fmt.Errorf("locating config file: %w", err)
	}
	key, err := configAPIKey(path)
	if err != nil {
		return "", err
	}
	if key != "" {
		return key, nil
	}

	return "", &noKeyError{configFile: path}
}

// configAPIKey reads the repeaterbook_api_key value from the config file at
// path. A missing file is not an error: it returns an empty key. If the file is
// group/world-readable it warns to stderr, since it may hold a secret.
func configAPIKey(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("checking config file %q: %w", path, err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		logError("warning: config file %s is group/world-readable; run 'chmod 600 %s'", path, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading config file %q: %w", path, err)
	}
	var c config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return "", fmt.Errorf("parsing config file %q: %w", path, err)
	}
	return strings.TrimSpace(c.RepeaterBookAPIKey), nil
}

// redactKey replaces any occurrence of the resolved API key in s with a
// placeholder, so request URLs/headers can be logged without leaking the secret.
func redactKey(s string) string {
	if repeaterBookAPIKey == "" {
		return s
	}
	return strings.ReplaceAll(s, repeaterBookAPIKey, "[REDACTED]")
}

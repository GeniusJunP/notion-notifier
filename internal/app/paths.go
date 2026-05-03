package app

import (
	"os"
	"path/filepath"
	"runtime"

	"notion-notifier/internal/logging"
)

// DefaultConfigDir returns the OS-standard configuration directory for notion-notifier.
//   - Linux/macOS: ~/.config/notion-notifier
//   - Windows:     %APPDATA%\notion-notifier
func DefaultConfigDir() string {
	base, err := os.UserConfigDir()
	if err != nil {
		return "."
	}
	return filepath.Join(base, "notion-notifier")
}

// DefaultDataDir returns the OS-standard data directory for notion-notifier.
//   - Linux:   ~/.local/share/notion-notifier
//   - macOS:   ~/Library/Application Support/notion-notifier
//   - Windows: %LOCALAPPDATA%\notion-notifier
func DefaultDataDir() string {
	switch runtime.GOOS {
	case "linux":
		// XDG_DATA_HOME or ~/.local/share
		if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
			return filepath.Join(dir, "notion-notifier")
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "."
		}
		return filepath.Join(home, ".local", "share", "notion-notifier")
	default:
		// macOS: ~/Library/Application Support/notion-notifier
		// Windows: %LOCALAPPDATA%\notion-notifier
		// Both handled by UserConfigDir which on macOS returns ~/Library/Application Support
		// and on Windows returns %APPDATA%. For data, we want %LOCALAPPDATA% on Windows.
		if runtime.GOOS == "windows" {
			if dir := os.Getenv("LOCALAPPDATA"); dir != "" {
				return filepath.Join(dir, "notion-notifier")
			}
		}
		base, err := os.UserConfigDir()
		if err != nil {
			return "."
		}
		return filepath.Join(base, "notion-notifier")
	}
}

// EnsureDefaults creates the config and data directories if they don't exist,
// and writes starter config/env files if missing.
func EnsureDefaults(cfgDir, dataDir string) error {
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}

	cfgFile := filepath.Join(cfgDir, "config.yaml")
	if _, err := os.Stat(cfgFile); os.IsNotExist(err) {
		if err := os.WriteFile(cfgFile, []byte(defaultConfigYAML), 0o644); err != nil {
			return err
		}
		logging.Info("INIT", "created default config: %s", cfgFile)
	}

	envFile := filepath.Join(cfgDir, "env.yaml")
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		if err := os.WriteFile(envFile, []byte(defaultEnvYAML), 0o644); err != nil {
			return err
		}
		logging.Info("INIT", "created default env: %s", envFile)
	}

	return nil
}

const defaultConfigYAML = `# notion-notifier configuration
# See docs/specification.md for full reference.

timezone: "Asia/Tokyo"

sync:
  check_interval: 15  # minutes

notifications:
  upcoming: []
  periodic: []

webhook:
  is_test: false
  notification:
    content_type: "application/json"
    payload_template: '{"content":{{json .Message}}}'
  internal_notification:
    content_type: "application/json"
    payload_template: '{"content":{{json .Message}}}'

calendar_sync:
  enabled: false
  interval_hours: 6
  lookahead_days: 30
`

const defaultEnvYAML = `# notion-notifier secrets / environment
# These values can also be set via environment variables.

notion:
  api_key: ""        # or NOTION_API_KEY env var
  database_id: ""    # or NOTION_DATABASE_ID env var

webhook:
  notification_url: ""           # or NOTIFICATION_WEBHOOK_URL env var
  internal_notification_url: ""  # or INTERNAL_NOTIFICATION_WEBHOOK_URL env var

# google:
#   calendar_id: ""
#   service_account_key_file: ""
#   service_account_key_json: ""  # or GOOGLE_SERVICE_ACCOUNT_KEY_JSON env var

server:
  port: 18080  # or APP_PORT env var
  # tls:
  #   cert_file: ""
  #   key_file: ""
`

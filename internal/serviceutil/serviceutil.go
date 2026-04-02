package serviceutil

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/kardianos/service"
)

const serviceName = "notion-notifier"

// HasUserServiceInstallation reports whether a user-level service definition
// exists for the current platform.
func HasUserServiceInstallation() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}

	var candidate string
	switch runtime.GOOS {
	case "linux":
		candidate = filepath.Join(home, ".config", "systemd", "user", serviceName+".service")
	case "darwin":
		candidate = filepath.Join(home, "Library", "LaunchAgents", serviceName+".plist")
	default:
		return false
	}

	_, err = os.Stat(candidate)
	return err == nil
}

// RestartHint returns the most accurate restart command for the current install.
func RestartHint() string {
	if HasUserServiceInstallation() {
		return "notion-notifier --user restart"
	}
	return "notion-notifier restart"
}

// Control runs a service action and transparently retries against a user-level
// service when a matching user service installation exists.
func Control(prg service.Interface, baseConfig *service.Config, action string, userRequested bool) error {
	s, err := service.New(prg, RuntimeConfig(baseConfig, userRequested))
	if err != nil {
		return fmt.Errorf("service init failed: %w", err)
	}

	err = service.Control(s, action)
	if err == nil || userRequested || !HasUserServiceInstallation() {
		return err
	}

	userService, createErr := service.New(prg, RuntimeConfig(baseConfig, true))
	if createErr != nil {
		return fmt.Errorf("%w (also failed to initialize user service: %v)", err, createErr)
	}
	if retryErr := service.Control(userService, action); retryErr != nil {
		return fmt.Errorf("%w (user service retry failed: %v)", err, retryErr)
	}
	return nil
}

// RuntimeConfig returns a copy of the service config tailored for the current
// invocation target.
func RuntimeConfig(base *service.Config, userService bool) *service.Config {
	cfg := *base
	if userService {
		cfg.Option = service.KeyValue{"UserService": true}
	} else {
		cfg.Option = nil
	}
	return &cfg
}

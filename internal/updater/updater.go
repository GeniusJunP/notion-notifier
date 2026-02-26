package updater

import (
	"fmt"
	"os"
	"strings"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

// Run checks GitHub for a newer release than the current version,
// downloads it, and replaces the currently running executable if an update is found.
// The repository parameter should be in the format "owner/repo", e.g. "GeniusJunP/notion-notifier".
func Run(currentVersion, repository string) error {
	v := strings.TrimPrefix(currentVersion, "v")
	if v == "dev" || v == "" {
		fmt.Println("Running a development version (dev). Skipping auto-update check.")
		return nil
	}

	parsedVersion, err := semver.Parse(v)
	if err != nil {
		return fmt.Errorf("failed to parse current version '%s': %w", v, err)
	}

	latest, found, err := selfupdate.DetectLatest(repository)
	if err != nil {
		return fmt.Errorf("error detecting latest release: %w", err)
	}
	if !found {
		return fmt.Errorf("no releases found in repository %s", repository)
	}

	if latest.Version.Equals(parsedVersion) || latest.Version.LT(parsedVersion) {
		fmt.Printf("Currently at the latest version (%s). No update needed.\n", currentVersion)
		return nil
	}

	fmt.Printf("New version available: %s\nDownloading from: %s\n", latest.Version, latest.AssetURL)
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not locate executable path: %w", err)
	}

	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		return fmt.Errorf("error occurred while updating binary: %w", err)
	}

	fmt.Printf("Successfully updated to version %s\n", latest.Version)
	fmt.Println("Please restart the service to apply changes. (e.g. `notion-notifier restart`)")
	return nil
}

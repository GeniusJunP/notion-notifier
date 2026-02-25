// Package timezone provides a shared timezone loading utility.
package timezone

import (
	"strings"
	"time"
)

// LoadOrLocal loads the named timezone, falling back to time.Local
// if the name is empty or invalid.
func LoadOrLocal(name string) *time.Location {
	if strings.TrimSpace(name) == "" {
		return time.Local
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.Local
	}
	return loc
}

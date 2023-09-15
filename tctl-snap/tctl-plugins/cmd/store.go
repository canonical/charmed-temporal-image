package cmd

import (
	"os"
	"path/filepath"
)

// TokenStoreLocation returns the location of the token store given a service
// name following the pattern of ~/.local/share/serviceName/tokens.
func TokenStoreLocation(serviceName string) string {
	if p := os.Getenv("XDG_DATA_HOME"); p != "" {
		return filepath.Join(p, serviceName, "tokens")
	}
	return filepath.Join(os.Getenv("HOME"), ".local", "share", serviceName, "tokens")
}

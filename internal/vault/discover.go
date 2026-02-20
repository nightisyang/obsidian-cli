package vault

import (
	"os"
	"path/filepath"
)

func DiscoverRoot(explicit string) (string, error) {
	if explicit != "" {
		return filepath.Abs(explicit)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	current := cwd
	for {
		candidate := filepath.Join(current, ".obsidian-cli.yaml")
		if _, statErr := os.Stat(candidate); statErr == nil {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return filepath.Abs(cwd)
}

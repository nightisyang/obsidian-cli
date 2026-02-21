package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/nightisyang/obsidian-cli/internal/app"
)

func verifyHashPrecondition(rt *app.Runtime, path, expectedHash string) error {
	expected := strings.TrimSpace(expectedHash)
	if expected == "" {
		return nil
	}
	n, err := rt.Backend.GetNote(rt.Context, path)
	if err != nil {
		return err
	}
	actual := hashString(n.Raw)
	if actual != expected {
		return fmt.Errorf("hash precondition failed for %s: expected %s got %s", path, expected, actual)
	}
	return nil
}

func hashString(raw string) string {
	digest := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(digest[:])
}

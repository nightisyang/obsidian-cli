package cmd

import (
	"os"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/index"
)

type operationMetadata struct {
	GeneratedAt        string `json:"generated_at"`
	SourceFileMTimeMax string `json:"source_file_mtime_max,omitempty"`
	CacheStatus        string `json:"cache_status,omitempty"`
	Truncated          bool   `json:"truncated"`
	Strict             bool   `json:"strict"`
}

func newOperationMetadata(strict bool) operationMetadata {
	return operationMetadata{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Strict:      strict,
	}
}

func sourceFileMTimeMax(vaultRoot string) (string, error) {
	files, err := index.ListMarkdownFiles(vaultRoot)
	if err != nil {
		return "", err
	}
	max := time.Time{}
	for _, abs := range files {
		info, statErr := os.Stat(abs)
		if statErr != nil {
			continue
		}
		if info.ModTime().After(max) {
			max = info.ModTime()
		}
	}
	if max.IsZero() {
		return "", nil
	}
	return max.UTC().Format(time.RFC3339), nil
}

package app

import (
	"context"

	"github.com/nightisyang/obsidian-cli/internal/backend"
	"github.com/nightisyang/obsidian-cli/internal/output"
	"github.com/nightisyang/obsidian-cli/internal/vault"
)

type Runtime struct {
	Context       context.Context
	Config        vault.Config
	VaultRoot     string
	ConfigPath    string
	ConfigSource  string
	RequestedMode string
	EffectiveMode string
	Printer       *output.Printer
	Backend       backend.Backend
}

package vault

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/nightisyang/obsidian-cli/internal/errs"
	"gopkg.in/yaml.v3"
)

type Config struct {
	VaultPath    string        `yaml:"vault_path" json:"vault_path"`
	ModeDefault  string        `yaml:"mode_default" json:"mode_default"`
	APIBaseURL   string        `yaml:"api_base_url" json:"api_base_url"`
	APITimeout   time.Duration `yaml:"-" json:"api_timeout"`
	TemplatesDir string        `yaml:"templates_dir" json:"templates_dir"`
	IndexDir     string        `yaml:"index_dir" json:"index_dir"`
}

type fileConfig struct {
	VaultPath    string `yaml:"vault_path"`
	ModeDefault  string `yaml:"mode_default"`
	APIBaseURL   string `yaml:"api_base_url"`
	APITimeout   string `yaml:"api_timeout"`
	TemplatesDir string `yaml:"templates_dir"`
	IndexDir     string `yaml:"index_dir"`
}

type Resolved struct {
	Config     Config
	VaultRoot  string
	ConfigPath string
	Source     string
}

func DefaultConfig() Config {
	return Config{
		VaultPath:    ".",
		ModeDefault:  "auto",
		APIBaseURL:   "https://127.0.0.1:27124",
		APITimeout:   5 * time.Second,
		TemplatesDir: ".obsidian/templates",
		IndexDir:     ".obsidian-cli-index",
	}
}

func ResolveConfig(explicitVault, explicitConfig string) (Resolved, error) {
	root, err := DiscoverRoot(explicitVault)
	if err != nil {
		return Resolved{}, errs.Wrap(errs.ExitConfig, "failed to discover vault", err)
	}

	cfg := DefaultConfig()
	var configPath string
	source := "defaults"

	if explicitConfig != "" {
		configPath = explicitConfig
		source = "explicit"
	} else {
		vaultCfg := filepath.Join(root, ".obsidian-cli.yaml")
		if _, statErr := os.Stat(vaultCfg); statErr == nil {
			configPath = vaultCfg
			source = "vault"
		} else {
			home, homeErr := os.UserHomeDir()
			if homeErr == nil {
				globalCfg := filepath.Join(home, ".obsidian-cli.yaml")
				if _, globalStatErr := os.Stat(globalCfg); globalStatErr == nil {
					configPath = globalCfg
					source = "global"
				}
			}
		}
	}

	if configPath != "" {
		loaded, loadErr := loadFileConfig(configPath)
		if loadErr != nil {
			return Resolved{}, errs.Wrap(errs.ExitConfig, "failed to load config", loadErr)
		}
		cfg = mergeConfig(cfg, loaded)
		if cfg.VaultPath != "" {
			if filepath.IsAbs(cfg.VaultPath) {
				root = filepath.Clean(cfg.VaultPath)
			} else {
				root = filepath.Clean(filepath.Join(filepath.Dir(configPath), cfg.VaultPath))
			}
		}
	}

	return Resolved{
		Config:     cfg,
		VaultRoot:  root,
		ConfigPath: configPath,
		Source:     source,
	}, nil
}

func WriteConfig(path string, cfg Config, force bool) error {
	if _, err := os.Stat(path); err == nil && !force {
		return errs.New(errs.ExitValidation, "config already exists; use --force to overwrite")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if mkErr := os.MkdirAll(filepath.Dir(path), 0o755); mkErr != nil {
		return mkErr
	}

	fc := fileConfig{
		VaultPath:    cfg.VaultPath,
		ModeDefault:  cfg.ModeDefault,
		APIBaseURL:   cfg.APIBaseURL,
		APITimeout:   cfg.APITimeout.String(),
		TemplatesDir: cfg.TemplatesDir,
		IndexDir:     cfg.IndexDir,
	}
	payload, err := yaml.Marshal(fc)
	if err != nil {
		return err
	}
	return os.WriteFile(path, payload, 0o644)
}

func loadFileConfig(path string) (fileConfig, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return fileConfig{}, err
	}
	var cfg fileConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return fileConfig{}, err
	}
	return cfg, nil
}

func mergeConfig(base Config, override fileConfig) Config {
	cfg := base
	if override.VaultPath != "" {
		cfg.VaultPath = override.VaultPath
	}
	if override.ModeDefault != "" {
		cfg.ModeDefault = override.ModeDefault
	}
	if override.APIBaseURL != "" {
		cfg.APIBaseURL = override.APIBaseURL
	}
	if override.APITimeout != "" {
		if d, err := time.ParseDuration(override.APITimeout); err == nil {
			cfg.APITimeout = d
		}
	}
	if override.TemplatesDir != "" {
		cfg.TemplatesDir = override.TemplatesDir
	}
	if override.IndexDir != "" {
		cfg.IndexDir = override.IndexDir
	}
	return cfg
}

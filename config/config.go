package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Providers map[string]string `mapstructure:"providers" yaml:"providers"`
	Exporters map[string]string `mapstructure:"exporters" yaml:"exporters"`
	Output    Output            `mapstructure:"output" yaml:"output"`
	Secrets   map[string]Secret `mapstructure:"secrets" yaml:"secrets"`
}

type Secret struct {
	Ref             string         `mapstructure:"ref" yaml:"ref"`
	Provider        string         `mapstructure:"provider" yaml:"provider"`
	ProviderOptions map[string]any `mapstructure:"provider_options" yaml:"provider_options"`
}

type Output struct {
	Type     string         `mapstructure:"type" yaml:"type"`
	Template string         `mapstructure:"template" yaml:"template"`
	Options  map[string]any `mapstructure:"options" yaml:"options"`
}

// Load reads configuration using viper, applies defaults, and decodes into Config.
func Load() (Config, error) {
	v := viper.New()
	v.SetConfigName(".sfx")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	v.SetDefault("providers.file", "./bin/providers/file")
	v.SetDefault("providers.vault", "./bin/providers/vault")
	v.SetDefault("providers.sops", "./bin/providers/sops")
	v.SetDefault("providers.awssecrets", "./bin/providers/awssecrets")
	v.SetDefault("providers.awsssm", "./bin/providers/awsssm")
	v.SetDefault("providers.gcpsecrets", "./bin/providers/gcpsecrets")
	v.SetDefault("providers.azurevault", "./bin/providers/azurevault")
	v.SetDefault("exporters.env", "./bin/exporters/env")
	v.SetDefault("output.type", "env")

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("SFX")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return Config{}, fmt.Errorf("no .sfx.yaml found in current directory")
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

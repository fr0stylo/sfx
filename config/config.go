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
	viper.SetConfigName(".sfx")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetDefault("providers.file", "./bin/providers/file")
	viper.SetDefault("providers.vault", "./bin/providers/vault")
	viper.SetDefault("providers.sops", "./bin/providers/sops")
	viper.SetDefault("providers.awssecrets", "./bin/providers/awssecrets")
	viper.SetDefault("providers.awsssm", "./bin/providers/awsssm")
	viper.SetDefault("providers.gcpsecrets", "./bin/providers/gcpsecrets")
	viper.SetDefault("providers.azurevault", "./bin/providers/azurevault")
	viper.SetDefault("exporters.env", "./bin/exporters/env")
	viper.SetDefault("exporters.tfvars", "./bin/exporters/tfvars")
	viper.SetDefault("exporters.template", "./bin/exporters/template")
	viper.SetDefault("exporters.shell", "./bin/exporters/shell")
	viper.SetDefault("exporters.k8ssecret", "./bin/exporters/k8ssecret")
	viper.SetDefault("exporters.ansible", "./bin/exporters/ansible")
	viper.SetDefault("output.type", "env")

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("SFX")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return Config{}, fmt.Errorf("no .sfx.yaml found in current directory")
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	return cfg, nil
}

package config

type Config struct {
	Providers map[string]string `yaml:"providers"`
	Output    Output            `yaml:"output"`
	Secrets   map[string]Secret `yaml:"secrets"`
}

type Secret struct {
	Ref      string `yaml:"ref"`
	Provider string `yaml:"provider"`
}

type Output struct {
	Type     string `yaml:"type"`
	Template string `yaml:"template"`
}

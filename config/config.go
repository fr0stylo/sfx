package config

type Config struct {
	Providers map[string]string `yaml:"providers"`
	Output    Output            `yaml:"output"`
	Secrets   map[string]Secret `yaml:"secrets"`
}

type Secret struct {
	Ref             string `yaml:"ref"`
	Provider        string `yaml:"provider"`
	ProviderOptions any    `yaml:"provider_options"`
}

type Output struct {
	Type          string `yaml:"type"`
	Template      string `yaml:"template"`
	OutputOptions []byte `yaml:"output_options"`
}

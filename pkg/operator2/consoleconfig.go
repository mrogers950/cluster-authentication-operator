package operator2

// ConsoleConfig is the top-level console configuration.
type ConsoleConfig struct {
	Customization `yaml:"customization"`
}

// Customization holds configuration such as what logo to use.
type Customization struct {
	Branding string `yaml:"branding"`
}

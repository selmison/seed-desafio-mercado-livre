package cmd

// Config is used to configure the server
type Config struct {
}

// NewConfig translates a RuntimeConfig into a Config.
func NewConfig() (*Config, error) {
	return &Config{}, nil
}

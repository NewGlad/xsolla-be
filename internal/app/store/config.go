package store

// Config ...
type Config struct {
	DSN string `yaml:dsn`
}

// NewConfig ...
func NewConfig() *Config{
	return &Config{}
}

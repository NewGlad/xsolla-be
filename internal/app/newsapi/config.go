package newsapi

import (
	"io/ioutil"
	"path/filepath"

	"github.com/NewGlad/xsolla-be/internal/app/store"

	"gopkg.in/yaml.v2"
)

// APIConfig ...
type APIConfig struct {
	BindAddr     string        `yaml:"bind_addr"`
	LogLevel     string        `yaml:"log_level"`
	StoreConfig  *store.Config `yaml:"store"`
	SessionKey   string        `yaml:"session_key"`
	TopNewsLimit int           `yaml:"top_news_limit"`
}

// NewConfig ...
func NewConfig(path string) (*APIConfig, error) {
	filename, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var config APIConfig
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

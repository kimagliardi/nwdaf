package factory

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

var NwdafConfig *Config

type Config struct {
	Info          *Info          `yaml:"info"`
	Configuration *Configuration `yaml:"configuration"`
	Logger        *Logger        `yaml:"logger,omitempty"`
}

type Info struct {
	Version     string `yaml:"version,omitempty"`
	Description string `yaml:"description,omitempty"`
}

type Configuration struct {
	NwdafName        string            `yaml:"nwdafName"`
	Sbi              *Sbi              `yaml:"sbi"`
	ServiceNameList  []string          `yaml:"serviceNameList"`
	NrfUri           string            `yaml:"nrfUri"`
	PlmnList         []PlmnId          `yaml:"plmnList"`
	AnalyticsDelay   int               `yaml:"analyticsDelay,omitempty"`
	DataCollectionConfig *DataCollectionConfig `yaml:"dataCollection,omitempty"`
}

type Sbi struct {
	Scheme       string `yaml:"scheme"`
	RegisterIPv4 string `yaml:"registerIPv4,omitempty"`
	BindingIPv4  string `yaml:"bindingIPv4,omitempty"`
	Port         int    `yaml:"port,omitempty"`
	TLS          *TLS   `yaml:"tls,omitempty"`
}

type TLS struct {
	Key  string `yaml:"key,omitempty"`
	PEM  string `yaml:"pem,omitempty"`
}

type PlmnId struct {
	Mcc string `yaml:"mcc"`
	Mnc string `yaml:"mnc"`
}

type DataCollectionConfig struct {
	Enabled           bool     `yaml:"enabled"`
	CollectionPeriod  int      `yaml:"collectionPeriod"`
	TargetNFs         []string `yaml:"targetNFs"`
}

type Logger struct {
	Level string `yaml:"level,omitempty"`
	File  string `yaml:"file,omitempty"`
}

func InitConfigFactory(configPath string) error {
	if configPath == "" {
		configPath = "config/nwdafcfg.yaml"
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	config := &Config{}
	if err := yaml.Unmarshal(content, config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	NwdafConfig = config

	// Set defaults if not specified
	if config.Configuration.AnalyticsDelay == 0 {
		config.Configuration.AnalyticsDelay = 10
	}

	if config.Configuration.Sbi.Port == 0 {
		config.Configuration.Sbi.Port = 8000
	}

	if config.Configuration.Sbi.Scheme == "" {
		config.Configuration.Sbi.Scheme = "http"
	}

	return nil
}

func (c *Config) GetVersion() string {
	if c.Info != nil && c.Info.Version != "" {
		return c.Info.Version
	}
	return "1.0.0"
}

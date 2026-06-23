package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
}

type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

var cfg *Config

func Load(configFile string) error {
	v := viper.New()

	v.SetConfigName(".gitli")
	v.SetConfigType("yaml")

	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		v.AddConfigPath(home)
		v.AddConfigPath(".")
	}

	v.SetDefault("database.path", filepath.Join(homeDir(), ".gitli", "gitli.db"))

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	cfg = &Config{}
	return v.Unmarshal(cfg)
}

func Get() *Config {
	if cfg == nil {
		_ = Load("")
	}
	return cfg
}

func homeDir() string {
	home, _ := os.UserHomeDir()
	return home
}
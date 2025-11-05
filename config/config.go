package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Debug bool
		Db    string
	}

	Storage struct {
		AccessKey string
		SecretKey string
		Region    string
		Bucket    string
		Endpoint  string
		Cname     string
		Point     string
	}
}

func NewConfig() *Config {

	return &Config{}
}

func (cfg *Config) LoadConfig() error {

	viper.SetConfigFile("temp/config.toml")

	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %v", err)
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return fmt.Errorf("failed to unpack config: %v", err)
	}

	return nil
}

func setDefaults() {
	viper.SetDefault("APP.DEBUG", true)
	viper.SetDefault("APP.DB", "")
}

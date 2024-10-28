package utils

import (
	"github.com/spf13/viper"
)

var AppVersion = "wip"

type AppConfig struct {
	Port int
}

func LoadConfig() AppConfig {
	viper.AutomaticEnv()

	cfg := AppConfig{}

	viper.SetDefault("PORT", 9090)
	cfg.Port = viper.GetInt("PORT")

	return cfg
}

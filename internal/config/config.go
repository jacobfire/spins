package config

import (
	"github.com/spf13/viper"
	"strings"
)

func Init(configDir string) error {
	return Load(configDir)
}

func Load(configDir string) error {
	viper.AddConfigPath(configDir)
	viper.SetConfigName("local")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	return viper.ReadInConfig()
}

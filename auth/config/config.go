package config

import (
	"github.com/spf13/viper"
)

func ParseConfig(path string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

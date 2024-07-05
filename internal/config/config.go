package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port        int    `mapstructure:"PORT"`
	AppEnv      string `mapstructure:"APP_ENV"`

	DatabaseURI string `mapstructure:"DATABASE_URI"`
}

func ReadConfig() *Config {
	config := Config{}
	viper.SetConfigFile(".env")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Can't find config file: %v", err)
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	switch config.AppEnv {
	case "local":
		log.Println("Service is running on 'local' env")
	case "dev":
		log.Println("Service is running on 'dev' env")
	case "prod":
		log.Println("Service is running on 'prod' env")
	}

	return &config
}
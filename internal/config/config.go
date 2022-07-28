package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	JenkinsUsername string `mapstructure:"JENKINS_USERNAME"`
	JenkinsPassword string `mapstructure:"JENKINS_PASSWORD"`
}

func LoadConfig(path string) (c Config, err error) {
	viper.SetConfigType("env")
	viper.SetConfigFile(path) // path to look for the config file in

	err = viper.ReadInConfig() // Find and read the config file
	if err != nil {            // Handle errors reading the config file
		return
	}

	err = viper.Unmarshal(&c)
	return
}

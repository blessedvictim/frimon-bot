package config

import (
	"github.com/blessedvictim/frimon-bot/model"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	SlackToken string      `mapstructure:"slack_token"`
	Jobs       []model.Job `mapstructure:"jobs"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.Wrap(err, "cannot read config")
	}

	var cfg Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal config")
	}

	return &cfg, nil
}

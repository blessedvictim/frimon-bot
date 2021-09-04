package config

import (
	"bytes"

	"github.com/blessedvictim/frimon-bot/model"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	SlackToken string      `mapstructure:"slack_token"`
	Jobs       []model.Job `mapstructure:"jobs"`
}

var embedRawConfig []byte

func LoadConfig() (*Config, error) {
	var err error

	viper.SetConfigType("json")

	if embedRawConfig != nil {
		err = viper.ReadConfig(bytes.NewReader(embedRawConfig))
	} else {
		viper.SetConfigName("config")
		viper.AddConfigPath("./config")
		err = viper.ReadInConfig()
	}
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

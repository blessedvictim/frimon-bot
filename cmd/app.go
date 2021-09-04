package main

import (
	"net/http"

	"github.com/blessedvictim/frimon-bot/app"
	"github.com/blessedvictim/frimon-bot/config"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	slackClient := slack.New(cfg.SlackToken)
	_, err = slackClient.AuthTest()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	a := app.New(slackClient, http.DefaultClient, cfg)

	err = a.Run()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
}

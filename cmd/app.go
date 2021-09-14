package main

import (
	"net/http"
	"sync"

	"github.com/blessedvictim/frimon-bot/config"
	"github.com/blessedvictim/frimon-bot/modules/answer"
	"github.com/blessedvictim/frimon-bot/modules/schedule"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

type Module interface {
	Run() error // должен быть блокирующим вызовом
}

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

	slackClient2 := slack.New(cfg.SlackToken)
	_, err = slackClient2.AuthTest()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	var modules []Module
	modules = append(modules, schedule.New(slackClient, http.DefaultClient, cfg))
	modules = append(modules, answer.New(slackClient2))

	wg := sync.WaitGroup{}

	for _, m := range modules {
		module := m
		wg.Add(1)

		go func() {
			defer wg.Done()

			err := module.Run()
			if err != nil {
				log.Fatal().Err(err).Send()
			}
		}()
	}

	wg.Wait()
}

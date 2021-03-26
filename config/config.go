package config

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DiscordAuthToken         string
	DiscordBeansWebhookID    string
	DiscordBeansWebhookToken string
}

func ProvideConfig() Config {
	var conf Config

	err := envconfig.Process("cafebean", &conf)
	if err != nil {
		log.Fatal(err.Error())
	}

	return conf
}

var Options = ProvideConfig

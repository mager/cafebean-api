package config

import (
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Discord `yaml:"discord"`
}

type Discord struct {
	AuthToken string          `yaml:"auth_token"`
	Webhooks  DiscordWebhooks `yaml:"webhooks"`
}

type DiscordWebhooks struct {
	Beans DiscordWebhooksBeansConfig `yaml:"beans"`
}

type DiscordWebhooksBeansConfig struct {
	WebhookID string `yaml:"webhook_id"`
	Token     string `yaml:"token"`
}

func ProvideConfig() *Config {
	conf := &Config{}
	data, err := ioutil.ReadFile("config/base.yaml")
	if err != nil {
		panic(err)
	}

	data = []byte(os.ExpandEnv(string(data)))
	if err := yaml.Unmarshal(data, conf); err != nil {
		panic(err)
	}

	return conf
}

var Options = ProvideConfig

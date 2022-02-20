package common

import "github.com/caarlos0/env/v6"

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"  envDefault:"localhost:8080"`
	BaseURL       string `env:"BASE_URL" envDefault:"http://localhost:8080"`
}

var Cfg Config

func init() {
	err := env.Parse(&Cfg)
	if err != nil {
		panic(err)
	}
}

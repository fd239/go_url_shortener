package common

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"  envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

var Cfg Config

func init() {
	err := env.Parse(&Cfg)
	if err != nil {
		panic(err)
	}

	flag.StringVar(&Cfg.ServerAddress, "ServerAddress", Cfg.ServerAddress, "server address")
	flag.StringVar(&Cfg.BaseURL, "BaseURL", Cfg.BaseURL, "base url")
	flag.StringVar(&Cfg.FileStoragePath, "FileStoragePath", Cfg.FileStoragePath, "file storage path")
}

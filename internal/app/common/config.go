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

func InitConfig() {
	err := env.Parse(&Cfg)

	if err != nil {
		panic(err)
	}

	addrPointer := flag.String("a", Cfg.ServerAddress, "Server address")
	baseAddrPointer := flag.String("b", Cfg.BaseURL, "Base URL address")
	fileStoragePathPointer := flag.String("f", Cfg.FileStoragePath, "File storage path")

	flag.Parse()

	Cfg.ServerAddress = *addrPointer
	Cfg.BaseURL = *baseAddrPointer
	Cfg.FileStoragePath = *fileStoragePathPointer
}

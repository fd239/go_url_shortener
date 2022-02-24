package common

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"  envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

var Cfg Config

func init() {
	flag.StringVar(&Cfg.ServerAddress, "ServerAddress", os.Getenv("SERVER_ADDRESS"), "server address")
	flag.StringVar(&Cfg.BaseURL, "BaseURL", os.Getenv("BASE_URL"), "base url")
	flag.StringVar(&Cfg.FileStoragePath, "FileStoragePath", os.Getenv("FILE_STORAGE_PATH"), "file storage path")
}

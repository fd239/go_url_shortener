package common

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"  envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN" envDefault:"postgres://fd239:fd239@localhost:5432/short_url"` //envDefault:"postgres://fd239:fd239@localhost:5432/short_url"
}

var Cfg Config

var SecretKey = []byte("passphrasewhichneedstobe32bytes!")

func InitConfig() error {
	err := env.Parse(&Cfg)

	if err != nil {
		return err
	}

	addrPointer := flag.String("a", Cfg.ServerAddress, "Server address")
	baseAddrPointer := flag.String("b", Cfg.BaseURL, "Base URL address")
	fileStoragePathPointer := flag.String("f", Cfg.FileStoragePath, "File storage path")
	DatabaseDSNPointer := flag.String("d", Cfg.DatabaseDSN, "PostgreSQL database credentials")

	flag.Parse()

	Cfg.ServerAddress = *addrPointer
	Cfg.BaseURL = *baseAddrPointer
	Cfg.FileStoragePath = *fileStoragePathPointer
	Cfg.DatabaseDSN = *DatabaseDSNPointer

	return nil

}

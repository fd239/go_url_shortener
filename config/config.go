package config

import (
	"encoding/json"
	"flag"
	"github.com/caarlos0/env/v6"
	"io/ioutil"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"  envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"` //envDefault:"postgres://fd239:fd239@localhost:5432/short_url"
	UseTls          bool   `env:"USE_TLS" envDefault:"false"`
	JsonCfgFilePath string `env:"CONFIG" envDefault:"./config/cfg.json"`
}

type JsonConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database-dsn"`
	UseTls          bool   `json:"use_tls"`
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
	UseTls := flag.Bool("s", Cfg.UseTls, "TLS server")
	JsonCfgFilePath := flag.String("c", Cfg.JsonCfgFilePath, "JSON config")

	flag.Parse()

	Cfg.ServerAddress = *addrPointer
	Cfg.BaseURL = *baseAddrPointer
	Cfg.FileStoragePath = *fileStoragePathPointer
	Cfg.DatabaseDSN = *DatabaseDSNPointer
	Cfg.UseTls = *UseTls
	Cfg.JsonCfgFilePath = *JsonCfgFilePath

	if Cfg.JsonCfgFilePath == "" {
		return nil
	}

	bytes, err := ioutil.ReadFile(Cfg.JsonCfgFilePath)

	if err != nil {
		return err
	}

	var jsonCfg JsonConfig

	err = json.Unmarshal(bytes, &jsonCfg)
	if err != nil {
		return err
	}

	if Cfg.ServerAddress == "" {
		Cfg.ServerAddress = jsonCfg.ServerAddress
	}

	if Cfg.BaseURL == "" {
		Cfg.BaseURL = jsonCfg.BaseURL
	}

	if Cfg.FileStoragePath == "" {
		Cfg.FileStoragePath = jsonCfg.FileStoragePath
	}

	if Cfg.DatabaseDSN == "" {
		Cfg.DatabaseDSN = jsonCfg.DatabaseDSN
	}

	if !Cfg.UseTls {
		Cfg.UseTls = jsonCfg.UseTls
	}

	return nil
}

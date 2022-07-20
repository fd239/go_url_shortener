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
	UseTLS          bool   `env:"USE_TLS" envDefault:"false"`
	JSONCfgFilePath string `env:"CONFIG" envDefault:"./config/cfg.json"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET" envDefault:""`
}

type JSONConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database-dsn"`
	UseTLS          bool   `json:"use_tls"`
	TrustedSubnet   string `json:"trusted_subnet"`
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
	UseTLS := flag.Bool("s", Cfg.UseTLS, "TLS server")
	JSONCfgFilePath := flag.String("c", Cfg.JSONCfgFilePath, "JSON config")
	trustedSubnet := flag.String("t", Cfg.TrustedSubnet, "Trusted subnet")

	flag.Parse()

	Cfg.ServerAddress = *addrPointer
	Cfg.BaseURL = *baseAddrPointer
	Cfg.FileStoragePath = *fileStoragePathPointer
	Cfg.DatabaseDSN = *DatabaseDSNPointer
	Cfg.UseTLS = *UseTLS
	Cfg.JSONCfgFilePath = *JSONCfgFilePath
	Cfg.TrustedSubnet = *trustedSubnet

	if Cfg.JSONCfgFilePath == "" {
		return nil
	}

	bytes, err := ioutil.ReadFile(Cfg.JSONCfgFilePath)

	if err != nil {
		return err
	}

	var jsonCfg JSONConfig

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

	if !Cfg.UseTLS {
		Cfg.UseTLS = jsonCfg.UseTLS
	}

	if Cfg.TrustedSubnet == "" {
		Cfg.TrustedSubnet = jsonCfg.TrustedSubnet
	}

	return nil
}

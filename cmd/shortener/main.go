package main

import (
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/fd239/go_url_shortener/internal/app/server"
	"log"
)

func main() {
	err := common.InitConfig()

	if err != nil {
		log.Fatalf("Init config error: %s", err.Error())
	}

	s, err := server.NewServer(common.Cfg.ServerAddress, common.Cfg.BaseURL)
	if err != nil {
		log.Fatalf("Server start error: %s", err.Error())
	}

	log.Fatal(s.Start())
}

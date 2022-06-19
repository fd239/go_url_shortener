package main

import (
	"fmt"
	"github.com/fd239/go_url_shortener/internal/app/common"
	"github.com/fd239/go_url_shortener/internal/app/server"
	"log"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Printf("Build version: %v\n", buildVersion)
	fmt.Printf("Build date: %v\n", buildDate)
	fmt.Printf("Build commit: %v\n", buildCommit)

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

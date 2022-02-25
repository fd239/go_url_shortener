package main

import (
	"github.com/fd239/go_url_shortener/internal/app"
	"github.com/fd239/go_url_shortener/internal/app/common"
)

func main() {
	app.InitDB(*common.Cfg.FileStoragePath)
	app.ServerStart()
}

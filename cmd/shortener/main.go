package main

import (
	"github.com/fd239/go_url_shortener/internal/app"
	"github.com/fd239/go_url_shortener/internal/app/common"
)

func main() {
	common.InitConfig()
	app.InitDB()
	app.ServerStart()
}

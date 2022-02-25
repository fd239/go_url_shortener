package main

import (
	"github.com/fd239/go_url_shortener/internal/app"
)

func main() {
	app.InitDB()
	app.ServerStart()
}
